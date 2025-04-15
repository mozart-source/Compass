"""Todo operations for the MCP server."""

import logging
import json
import re
from typing import Dict, Any, Optional
from mcp.server.fastmcp import Context
from ai_services.llm.llm_service import LLMService
import httpx
from core.config import settings

# Configure logging
logger = logging.getLogger(__name__)

# Hardcoded JWT token for development - only used as fallback
DEV_JWT_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNDA4YjM4YmMtNWRlZS00YjA0LTlhMDYtZWE4MTk0OWJmNWMzIiwiZW1haWwiOiJhaG1lZEBnbWFpbC5jb20iLCJyb2xlcyI6WyJ1c2VyIl0sIm9yZ19pZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInBlcm1pc3Npb25zIjpbInRhc2tzOnJlYWQiLCJvcmdhbml6YXRpb25zOnJlYWQiLCJwcm9qZWN0czpyZWFkIiwidGFza3M6dXBkYXRlIiwidGFza3M6Y3JlYXRlIl0sImV4cCI6MTc0NjUwNDg1NiwibmJmIjoxNzQ2NDE4NDU2LCJpYXQiOjE3NDY0MTg0NTZ9.nUky6q0vPRnVYP9gTPIPaibNezB-7Sn-EgDZvlxU0_8"

# List of backend URLs to try
GO_BACKEND_URLS = [
    # Use settings for primary URL (uses proper Docker service names)
    settings.GO_BACKEND_URL,
    "http://api:8000",         # Docker service name fallback
    "http://localhost:8000"    # Fallback for local development
]

# Start with the primary URL from settings
GO_BACKEND_URL = settings.GO_BACKEND_URL


async def try_backend_urls(client_func, endpoint: str, **kwargs) -> Dict[str, Any]:
    """Try to connect to multiple backend URLs in sequence."""
    global GO_BACKEND_URL

    errors = []
    timeout = httpx.Timeout(10.0, connect=5.0)

    logger.info(
        f"[CONNECTION] Trying to connect to endpoint {endpoint} with {len(GO_BACKEND_URLS)} URLs")
    logger.info(f"[CONNECTION] Available URLs: {GO_BACKEND_URLS}")

    initial_url = GO_BACKEND_URL
    logger.info(f"[CONNECTION] Starting with URL: {initial_url}")

    for base_url in GO_BACKEND_URLS:
        full_url = f"{base_url}{endpoint}"
        logger.info(f"[CONNECTION] Attempting connection to: {full_url}")

        try:
            async with httpx.AsyncClient(timeout=timeout) as client:
                logger.info(
                    f"[CONNECTION] Sending {client_func.__name__} request to {full_url}")
                response = await client_func(client, full_url, **kwargs)
                logger.info(
                    f"[CONNECTION] Received response from {full_url}: status={response.status_code}")
                response.raise_for_status()

                previous_url = GO_BACKEND_URL
                GO_BACKEND_URL = base_url
                logger.info(
                    f"[CONNECTION] CONNECTION SUCCESS! {base_url} is working")
                logger.info(
                    f"[CONNECTION] Updated primary backend URL from {previous_url} to {GO_BACKEND_URL}")

                try:
                    result = response.json()
                    logger.info(
                        f"[CONNECTION] Successfully parsed JSON response from {base_url}")
                    return result
                except Exception as json_error:
                    logger.warning(
                        f"[CONNECTION] Response not JSON: {str(json_error)}")
                    return {"status": "success", "message": response.text}
        except Exception as e:
            logger.warning(
                f"[CONNECTION] Failed to connect to {base_url}: {str(e)}")
            errors.append({"url": base_url, "error": str(e),
                          "type": "connection_error"})

    error_msg = f"Failed to connect to any backend URL: {[e['url'] for e in errors]}"
    logger.error(
        f"[CONNECTION] ALL CONNECTION ATTEMPTS FAILED. Tried URLs: {GO_BACKEND_URLS}")
    logger.error(f"[CONNECTION] {error_msg}")
    logger.error(f"[CONNECTION] Last working URL was: {initial_url}")

    return {
        "status": "error",
        "error": error_msg,
        "type": "connection_error",
        "details": errors
    }


async def smart_update_todo(
    ctx: Context,
    edit_request: str,
    authorization: Optional[str] = None,
    user_id: Optional[str] = None
) -> Dict[str, Any]:
    """Intelligently update a todo item based on the user's description.

    Args:
        edit_request: User's request describing what to edit (e.g., "change the due date of my shopping task to tomorrow")
        authorization: Optional authorization token (Bearer token)
        user_id: Optional user ID (will be extracted from token if not provided)

    Returns:
        The updated todo item
    """
    try:
        logger.info(f"Smart update request: '{edit_request}'")
        await ctx.info(f"Processing todo update request: '{edit_request}'")

        # Get auth token from parameter or fall back to default
        auth_token = None
        if authorization and authorization.startswith("Bearer ") and authorization != "Bearer undefined" and authorization != "Bearer null":
            auth_token = authorization
            logger.info("Using provided authorization token")
        else:
            auth_token = f"Bearer {DEV_JWT_TOKEN}"
            logger.info(
                f"Using DEV_JWT_TOKEN for authorization: {auth_token[:20]}...")

        headers = {
            "Content-Type": "application/json",
            "Authorization": auth_token
        }

        async def get_func(client, url, **kwargs):
            return await client.get(url, **kwargs)

        todos_result = await try_backend_urls(
            get_func,
            "/api/todo-lists",
            headers=headers
        )

        if todos_result.get("status") == "error":
            error_msg = todos_result.get("error", "Failed to fetch todos")
            logger.error(f"Failed to fetch todos: {error_msg}")
            await ctx.error(f"Failed to fetch todos: {error_msg}")
            return {"status": "error", "error": error_msg}

        logger.info(f"Received todos data structure: {type(todos_result)}")
        if "data" in todos_result:
            logger.info(
                f"Todos data contains 'data' field: {type(todos_result['data'])}")

        todos_context = todos_result

        llm_service = LLMService()

        prompt = f"""
        User wants to edit a todo with this request: "{edit_request}"
        
        Here are the user's current todos:
        {json.dumps(todos_context, indent=2)}
        
        Based on the user's request and their todos, identify:
        1. Which todo needs to be edited (provide the todo_id)
        2. What specific fields need to be updated
        
        Return ONLY a valid JSON object without explanation. The JSON should contain:
        - todo_id: The UUID of the todo to update
        - title: New title (if the user wants to change it)
        - description: New description (if the user wants to change it)
        - due_date: New due date in YYYY-MM-DD format (if the user wants to change it)
        - priority: New priority as "high", "medium", or "low" (if the user wants to change it)
        - status: New status as "pending", "in_progress", or "archived" (if the user wants to change it)
        - is_completed: Boolean true/false (if the user wants to change completion status)
        
        Include ONLY the fields that need to be updated.
        """

        analysis_response = await llm_service.generate_response(
            prompt=prompt,
            context={
                "system_prompt": "You are a helpful assistant that analyzes todo items and user requests. Identify which todo the user wants to edit and what changes they want to make. Respond with a JSON object that includes todo_id and ONLY the fields that need to be updated."
            },
            stream=False
        )

        # Handle the response properly based on its type
        if hasattr(analysis_response, 'get'):
            response_text = analysis_response.get("text", "")
        elif hasattr(analysis_response, '__dict__'):
            response_text = getattr(
                analysis_response, 'text', str(analysis_response))
        else:
            response_text = str(analysis_response)

        logger.info(f"LLM response: {response_text[:200]}...")

        update_info = None
        try:
            json_match = re.search(
                r'```json\s*(.*?)\s*```', response_text, re.DOTALL)
            if json_match:
                json_text = json_match.group(1)
                logger.info(
                    f"Extracted JSON from code block: {json_text[:200]}...")
            else:
                json_pattern = r'({[^{]*"todo_id"[^}]*})'
                json_match = re.search(json_pattern, response_text, re.DOTALL)
                if json_match:
                    json_text = json_match.group(1)
                    logger.info(
                        f"Extracted JSON using pattern: {json_text[:200]}...")
                else:
                    json_text = response_text.strip()
                    logger.info(
                        f"Using full response as JSON: {json_text[:200]}...")

            json_text = re.sub(r'[^\x20-\x7E]', '', json_text)
            update_info = json.loads(json_text)
            logger.info(f"Parsed update info: {update_info}")
        except Exception as e:
            error_msg = f"Failed to parse LLM response: {str(e)}"
            logger.error(error_msg)
            await ctx.error(error_msg)
            return {"status": "error", "error": error_msg}

        if not update_info or "todo_id" not in update_info:
            error_msg = "AI couldn't identify which todo to update"
            logger.error(error_msg)
            await ctx.error(error_msg)
            return {"status": "error", "error": error_msg}

        todo_id = update_info.pop("todo_id", None)
        if not todo_id:
            error_msg = "No todo ID provided for update"
            logger.error(error_msg)
            await ctx.error(error_msg)
            return {"status": "error", "error": error_msg}

        is_completion_update = "is_completed" in update_info
        completion_value = update_info.pop("is_completed", None)

        if is_completion_update:
            async def patch_func(client, url, **kwargs):
                response = await client.patch(url, **kwargs)
                logger.info(f"PATCH response status: {response.status_code}")
                logger.info(f"PATCH response headers: {response.headers}")
                try:
                    logger.info(f"PATCH response text: {response.text[:500]}")
                except:
                    logger.info("Could not retrieve response text")
                return response

            completion_endpoint = f"/api/todos/{todo_id}/complete" if completion_value else f"/api/todos/{todo_id}/uncomplete"

            logger.info(
                f"Using dedicated completion endpoint: {completion_endpoint}")

            update_result = await try_backend_urls(
                patch_func,
                completion_endpoint,
                headers=headers,
                json={}
            )

            logger.info(
                f"Completion update result: {json.dumps(update_result, default=str)[:200]}...")

            if update_info:
                logger.info(
                    f"Additional fields to update: {update_info}. Proceeding with general update.")
            else:
                if update_result.get("status") != "error":
                    await ctx.info(f"Todo {todo_id} completion status updated successfully")
                    return {
                        "status": "success",
                        "message": f"Todo marked as {'completed' if completion_value else 'uncompleted'} successfully",
                        "content": update_result.get("content", {}) or update_result.get("data", {})
                    }
                else:
                    error_msg = update_result.get(
                        "error", "Unknown error updating todo completion status")
                    logger.error(
                        f"Error updating todo completion status: {error_msg}")
                    await ctx.error(f"Error updating todo completion status: {error_msg}")
                    return {
                        "status": "error",
                        "error": error_msg,
                        "type": "api_error"
                    }

        update_data = {}

        if "title" in update_info:
            update_data["title"] = update_info["title"]

        if "description" in update_info:
            update_data["description"] = update_info["description"]

        if "status" in update_info:
            update_data["status"] = update_info["status"]

        if "priority" in update_info:
            update_data["priority"] = update_info["priority"]

        if "due_date" in update_info:
            update_data["due_date"] = update_info["due_date"]

        if update_data:
            logger.info(f"Prepared update data: {update_data}")
            logger.info(f"Making PUT request to: /api/todos/{todo_id}")

            async def put_func(client, url, **kwargs):
                response = await client.put(url, **kwargs)
                logger.info(f"PUT response status: {response.status_code}")
                logger.info(f"PUT response headers: {response.headers}")
                try:
                    logger.info(f"PUT response text: {response.text[:500]}")
                except:
                    logger.info("Could not retrieve response text")
                return response

            update_result = await try_backend_urls(
                put_func,
                f"/api/todos/{todo_id}",
                headers=headers,
                json=update_data
            )

            logger.info(
                f"Update result: {json.dumps(update_result, default=str)[:200]}...")

            if update_result.get("status") != "error":
                await ctx.info(f"Todo {todo_id} updated successfully")
                return {
                    "status": "success",
                    "message": "Todo updated successfully",
                    "content": update_result.get("content", {}) or update_result.get("data", {})
                }
            else:
                error_msg = update_result.get(
                    "error", "Unknown error updating todo")
                logger.error(f"Error updating todo: {error_msg}")
                await ctx.error(f"Error updating todo: {error_msg}")
                return {
                    "status": "error",
                    "error": error_msg,
                    "type": "api_error"
                }
        else:
            return {
                "status": "success",
                "message": "Todo updated successfully",
            }

    except Exception as e:
        error_msg = f"Error in smart todo update: {str(e)}"
        logger.error(error_msg)
        await ctx.error(error_msg)
        return {"status": "error", "error": error_msg, "type": "api_error"}
