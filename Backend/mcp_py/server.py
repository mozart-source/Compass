from fastapi import FastAPI, Request, HTTPException, Header
from mcp.server.fastmcp import FastMCP, Context
from starlette.routing import Mount
from typing import Dict, Any, Optional, AsyncIterator, Union, AsyncGenerator, List
from fastapi.responses import StreamingResponse
import logging
import httpx
import os
import json
import sys
import asyncio
from mcp.types import (
    InitializeResult,
    ServerCapabilities,
    Implementation,
    ToolsCapability,
    LoggingCapability
)
import sys
from core.config import settings
from orchestration.ai_orchestrator import AIOrchestrator
from orchestration.todo_operations import smart_update_todo
import uuid
from data_layer.cache.ai_cache_manager import AICacheManager
from ai_services.llm.llm_service import LLMService
from datetime import datetime, timezone

# Hardcoded JWT token for development - only used as fallback
DEV_JWT_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiNDA4YjM4YmMtNWRlZS00YjA0LTlhMDYtZWE4MTk0OWJmNWMzIiwiZW1haWwiOiJhaG1lZEBnbWFpbC5jb20iLCJyb2xlcyI6WyJ1c2VyIl0sIm9yZ19pZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMCIsInBlcm1pc3Npb25zIjpbInRhc2tzOnJlYWQiLCJvcmdhbml6YXRpb25zOnJlYWQiLCJwcm9qZWN0czpyZWFkIiwidGFza3M6dXBkYXRlIiwidGFza3M6Y3JlYXRlIl0sImV4cCI6MTc0NjUwNDg1NiwibmJmIjoxNzQ2NDE4NDU2LCJpYXQiOjE3NDY0MTg0NTZ9.nUky6q0vPRnVYP9gTPIPaibNezB-7Sn-EgDZvlxU0_8"

print("PYTHONPATH:", sys.path)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout),
        logging.FileHandler('mcp_server.log')
    ]
)
logger = logging.getLogger(__name__)

# Initialize FastAPI app
app = FastAPI()

# Initialize FastMCP server with explicit file path handling
current_dir = os.path.dirname(os.path.abspath(__file__))
mcp = FastMCP(
    name="compass",
    version="1.0.0",
    endpoint="/mcp",
    prefix="/mcp",
    instructions="COMPASS AI Service MCP Server"
)

# Add a diagnostic endpoint to check registered tools


@app.get("/mcp-diagnostic")
async def mcp_diagnostic():
    """Diagnostic endpoint to verify MCP server configuration and tool registration."""
    try:
        # Get available tools - using proper public properties instead of internal ones
        registered_tools = []

        # Use reflection to safely gather tool information
        tools_dict = {}
        for attr_name in dir(mcp):
            if attr_name.startswith("_tool_"):
                tool_name = attr_name[6:]  # Remove "_tool_" prefix
                tool_func = getattr(mcp, attr_name)
                tool_description = getattr(
                    tool_func, "__doc__", "No description")
                tools_dict[tool_name] = {
                    "name": tool_name,
                    "description": tool_description
                }

        # Return diagnostic information
        return {
            "status": "running",
            "mcp_name": mcp.name,
            "mcp_version": getattr(mcp, "version", "1.0.0"),
            "tool_count": len(tools_dict),
            "registered_tools": list(tools_dict.values()),
            "backend_urls": GO_BACKEND_URLS,
            "current_backend_url": GO_BACKEND_URL,
        }
    except Exception as e:
        logger.error(f"Error in diagnostic endpoint: {str(e)}")
        return {
            "status": "error",
            "error": str(e),
            "mcp_initialized": bool(mcp),
            "backend_urls": GO_BACKEND_URLS
        }

# Define multiple backend URLs to try for Docker/non-Docker environments
GO_BACKEND_URLS = [
    settings.GO_BACKEND_URL,  # Use settings for primary URL
    "http://api:8000",         # Docker service name
    "http://localhost:8000",
    "http://localhost:8081/api"     # Fallback for local development
]

# Start with the primary URL from settings
GO_BACKEND_URL = settings.GO_BACKEND_URL
logger.info(f"Primary backend URL: {GO_BACKEND_URL}")
logger.info(f"Available backend URLs: {GO_BACKEND_URLS}")
HEADERS = {"Content-Type": "application/json"}


# Helper function to try multiple backend URLs
async def try_backend_urls(client_func, endpoint: str, **kwargs) -> Dict[str, Any]:
    """Try to connect to multiple backend URLs in sequence."""
    global GO_BACKEND_URL

    errors = []
    # Increase timeout for Docker networking
    timeout = httpx.Timeout(10.0, connect=5.0)

    logger.info(
        f"[CONNECTION] Trying to connect to endpoint {endpoint} with {len(GO_BACKEND_URLS)} URLs")
    logger.info(f"[CONNECTION] Available URLs: {GO_BACKEND_URLS}")

    # Track initial URL for logging purposes
    initial_url = GO_BACKEND_URL
    logger.info(f"[CONNECTION] Starting with URL: {initial_url}")

    for base_url in GO_BACKEND_URLS:
        full_url = f"{base_url}{endpoint}"
        logger.info(f"[CONNECTION] Attempting connection to: {full_url}")

        try:
            async with httpx.AsyncClient(timeout=timeout) as client:
                # Call the appropriate HTTP method function
                logger.info(
                    f"[CONNECTION] Sending {client_func.__name__} request to {full_url}")
                response = await client_func(client, full_url, **kwargs)
                logger.info(
                    f"[CONNECTION] Received response from {full_url}: status={response.status_code}")
                response.raise_for_status()

                # If successful, update the global URL for future requests
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
                    # Handle case where response isn't valid JSON
                    logger.warning(
                        f"[CONNECTION] Response not JSON: {str(json_error)}")
                    return {"status": "success", "message": response.text}
        except httpx.ConnectError as e:
            # Connection errors are expected when trying different URLs
            logger.warning(
                f"[CONNECTION] Connection error to {base_url}: {str(e)}")
            errors.append({"url": base_url, "error": str(e),
                          "type": "connection_error"})
        except httpx.TimeoutException as e:
            # Timeout errors
            logger.warning(
                f"[CONNECTION] Timeout connecting to {base_url}: {str(e)}")
            errors.append(
                {"url": base_url, "error": str(e), "type": "timeout"})
        except httpx.HTTPStatusError as e:
            # HTTP status errors (4xx, 5xx)
            logger.warning(
                f"[CONNECTION] HTTP error from {base_url}: {e.response.status_code}")
            errors.append({"url": base_url, "error": f"HTTP {e.response.status_code}",
                          "type": "http_error", "status": e.response.status_code})
        except Exception as e:
            # Other unexpected errors
            logger.warning(
                f"[CONNECTION] Failed to connect to {base_url}: {str(e)}")
            errors.append(
                {"url": base_url, "error": str(e), "type": "unexpected"})

    # If we get here, all URLs failed
    error_msg = f"Failed to connect to any backend URL: {[e['url'] for e in errors]}"
    logger.error(
        f"[CONNECTION] ALL CONNECTION ATTEMPTS FAILED. Tried URLs: {GO_BACKEND_URLS}")
    logger.error(f"[CONNECTION] {error_msg}")
    logger.error(f"[CONNECTION] Last working URL was: {initial_url}")

    # Return a structured error response instead of raising an exception
    # This allows the client to handle the error more gracefully
    return {
        "status": "error",
        "error": error_msg,
        "type": "connection_error",
        "details": errors
    }


@mcp.tool("create.user")
async def create_user(user_data: Dict[str, Any], ctx: Context) -> Dict[str, Any]:
    """Create a new user in the system.

    Args:
        user_data: Dictionary containing user information
            - email: str
            - username: str 
            - password: str
            - firstName: str (will be converted to first_name)
            - lastName: str (will be converted to last_name)
            - phoneNumber: str (will be converted to phone_number)
            - timezone: str
            - locale: str
    """
    try:
        # Transform camelCase to snake_case for the Go backend
        transformed_data = {
            "email": user_data.get("email"),
            "username": user_data.get("username"),
            "password": user_data.get("password"),
            "first_name": user_data.get("firstName"),
            "last_name": user_data.get("lastName"),
            "phone_number": user_data.get("phoneNumber"),
            "timezone": user_data.get("timezone"),
            "locale": user_data.get("locale")
        }

        # Remove None values
        transformed_data = {k: v for k,
                            v in transformed_data.items() if v is not None}

        async def post_func(client, url, **kwargs):
            return await client.post(url, **kwargs)

        return await try_backend_urls(
            post_func,
            "/api/users/register",
            json=transformed_data,
            headers=HEADERS
        )
    except Exception as e:
        await ctx.error(f"Failed to create user: {str(e)}")
        raise


@mcp.tool("check.health")
async def check_health(ctx: Context) -> Dict[str, Any]:
    """Check the health status of the system.

    Returns:
        Dict containing health check information
    """
    try:
        async def get_func(client, url, **kwargs):
            return await client.get(url, **kwargs)

        return await try_backend_urls(
            get_func,
            "/health",
            headers=HEADERS
        )
    except Exception as e:
        await ctx.error(f"Health check failed: {str(e)}")
        raise


@mcp.tool()
async def create_task(task_data: Dict[str, Any], ctx: Context) -> Dict[str, Any]:
    """Create a new task.

    Args:
        task_data: Dictionary containing task information
            - title: str
            - description: str
            - due_date: str (optional)
            - priority: str (optional)
    """
    try:
        async def post_func(client, url, **kwargs):
            return await client.post(url, **kwargs)

        return await try_backend_urls(
            post_func,
            "/api/tasks",
            json=task_data,
            headers=HEADERS
        )
    except Exception as e:
        await ctx.error(f"Failed to create task: {str(e)}")
        raise


@mcp.tool()
async def get_tasks(
    ctx: Context,
    user_id: str,
    authorization: Optional[str] = None,
    start_date: Optional[str] = None,
    end_date: Optional[str] = None,
    status: Optional[str] = None,
    priority: Optional[str] = None,
    project_id: Optional[str] = None
) -> Dict[str, Any]:
    """
    Retrieves a list of tasks for a given user, with optional filters.
    This tool fetches tasks where the user is either the creator or the assignee and
    filters them by the provided date range.
    """
    logger.info(
        f"get_tasks called for user_id: {user_id} with date range: {start_date} to {end_date}")

    auth_token = None
    if authorization and authorization.startswith("Bearer ") and authorization != "Bearer undefined" and authorization != "Bearer null":
        auth_token = authorization
    else:
        auth_token = f"Bearer {DEV_JWT_TOKEN}"

    headers = {"Content-Type": "application/json", "Authorization": auth_token}

    all_tasks = {}

    async def fetch_paged_tasks(base_params: Dict[str, Any]):
        page = 0
        page_size = 100
        while True:
            params = base_params.copy()
            params["page"] = str(page)
            params["page_size"] = str(page_size)

            async def get_func(client, url, **kwargs):
                return await client.get(url, **kwargs)

            result = await try_backend_urls(
                get_func,
                "/api/tasks",
                headers=headers,
                params=params
            )

            if result.get("status") == "error":
                logger.error(
                    f"Failed to fetch tasks with params {params}: {result.get('error')}")
                break

            data = result.get("data", {})
            tasks = data.get("tasks", [])

            for task in tasks:
                all_tasks[task['id']] = task

            if len(tasks) < page_size:
                break

            page += 1

    common_params = {}
    if status:
        common_params['status'] = status
    if priority:
        common_params['priority'] = priority
    if project_id:
        common_params['project_id'] = project_id

    # Fetch tasks created by user
    logger.info(f"Fetching tasks created by user {user_id}")
    creator_params = {"creator_id": user_id, **common_params}
    await fetch_paged_tasks(creator_params)

    # Fetch tasks assigned to user
    logger.info(f"Fetching tasks assigned to user {user_id}")
    assignee_params = {"assignee_id": user_id, **common_params}
    await fetch_paged_tasks(assignee_params)

    tasks_list = list(all_tasks.values())

    # Filter by date range if provided
    if start_date and end_date:
        try:
            start_dt_naive = datetime.fromisoformat(
                start_date.replace('Z', '+00:00'))
            start_dt = start_dt_naive.replace(
                tzinfo=timezone.utc) if start_dt_naive.tzinfo is None else start_dt_naive

            end_dt_naive = datetime.fromisoformat(
                end_date.replace('Z', '+00:00'))
            end_dt = end_dt_naive.replace(
                tzinfo=timezone.utc) if end_dt_naive.tzinfo is None else end_dt_naive

            filtered_tasks = []
            for task in tasks_list:
                task_start_date_str = task.get("start_date")
                task_due_date_str = task.get("due_date")

                task_date_to_check = None
                if task_start_date_str:
                    try:
                        task_date_to_check = datetime.fromisoformat(
                            task_start_date_str.replace('Z', '+00:00'))
                    except ValueError:
                        logger.warning(
                            f"Could not parse task start_date: {task_start_date_str}")

                if not task_date_to_check and task_due_date_str:
                    try:
                        task_date_to_check = datetime.fromisoformat(
                            task_due_date_str.replace('Z', '+00:00'))
                    except ValueError:
                        logger.warning(
                            f"Could not parse task due_date: {task_due_date_str}")

                if task_date_to_check and start_dt <= task_date_to_check <= end_dt:
                    filtered_tasks.append(task)

            tasks_list = filtered_tasks
            logger.info(
                f"Filtered tasks by date range. Found {len(tasks_list)} tasks.")

        except ValueError as e:
            logger.error(f"Invalid date format for filtering: {e}")
            await ctx.error(f"Invalid date format provided: {e}")
            return {"status": "error", "error": f"Invalid date format: {e}"}

    logger.info(f"get_tasks finished. Returning {len(tasks_list)} tasks.")
    return {
        "status": "success",
        "tasks": tasks_list,
        "total": len(tasks_list)
    }


@mcp.tool()
async def create_project(project_data: Dict[str, Any], ctx: Context) -> Dict[str, Any]:
    """Create a new project.

    Args:
        project_data: Dictionary containing project information
            - name: str
            - description: str
            - start_date: str (optional)
            - end_date: str (optional)
    """
    try:
        async def post_func(client, url, **kwargs):
            return await client.post(url, **kwargs)

        return await try_backend_urls(
            post_func,
            "/api/v1/projects",
            json=project_data,
            headers=HEADERS
        )
    except Exception as e:
        await ctx.error(f"Failed to create project: {str(e)}")
        raise


@mcp.tool()
async def get_projects(
    ctx: Context,
    user_id: str,
    authorization: Optional[str] = None
) -> Dict[str, Any]:
    """
    (Placeholder) Retrieves a list of projects for a given user.
    Currently returns an empty list.
    """
    logger.warning(
        "MCP tool 'get_projects' is a placeholder and will return an empty list.")
    return {
        "status": "success",
        "projects": [],
        "total": 0
    }


@mcp.tool("entity.create")
async def create_entity(
    ctx: Context,
    prompt: str,
    domain: Optional[str] = None
) -> Dict[str, Any]:
    """Create an entity from a prompt."""
    try:
        logger.info(
            f"Creating entity from prompt in domain: {domain or 'default'}")
        return {
            "entity_id": "123456",
            "response": f"Created entity from: {prompt[:20]}...",
            "intent": "create",
            "target": domain or "default",
            "description": "Entity created from description",
            "rag_used": False,
            "cached": False,
            "confidence": 0.9
        }
    except Exception as e:
        logger.error(f"Error creating entity: {str(e)}")
        raise


@mcp.tool("user.getInfo")
async def get_user_info(
    ctx: Context,
    user_id: str,
    authorization: Optional[str] = None
) -> Dict[str, Any]:
    """
    Retrieves user information from the Go backend.
    """
    logger.info(f"Received request for user.getInfo for user_id: {user_id}")
    try:
        request_id = str(uuid.uuid4())
        headers = {
            "X-Internal-Service": "mcp-server",
            "X-Request-ID": request_id,
            "User-Agent": "MCP-Server/1.0"
        }
        if authorization:
            headers["Authorization"] = authorization

        async def get_func(client, url, **kwargs):
            return await client.get(url, headers=headers, timeout=10.0, **kwargs)

        user_info_url = "/api/users/profile"
        response_data = await try_backend_urls(get_func, user_info_url)

        return response_data

    except Exception as e:
        logger.error(
            f"Error in get_user_info for user {user_id}: {e}", exc_info=True)
        return {
            "user_id": user_id,
            "name": "Unknown User",
            "email": "unknown@example.com",
            "error": str(e)
        }


@mcp.tool("get_items")
async def get_items(
    ctx: Context,
    item_type: str,  # "todos" or "habits"
    status: Optional[str] = None,
    priority: Optional[str] = None,
    authorization: Optional[str] = None,
    page: Optional[int] = None,
    page_size: Optional[int] = None,
    user_id: Optional[str] = None  # Moved to end to de-emphasize
) -> Dict[str, Any]:
    """Get/List/Show items (todos or habits) with optional filters.

    Args:
        item_type: Type of items to retrieve ("todos" or "habits")
        status: Optional status to filter by
        priority: Optional priority to filter by
        authorization: Optional authorization token (Bearer token)
        page: Optional page number for pagination
        page_size: Optional page size for pagination
        user_id: Optional user ID to filter by (not required - will use token's user)

    Returns:
        The list of items matching the filters
    """
    try:
        logger.info(
            f"get_items called with: type={item_type}, status={status}, priority={priority}")
        await ctx.info(f"get_items called with: type={item_type}, status={status}, priority={priority}")

        # Validate item type
        if item_type not in ["todos", "habits"]:
            error_msg = f"Invalid item type: {item_type}. Must be 'todos' or 'habits'"
            logger.error(error_msg)
            return {"status": "error", "error": error_msg, "type": "validation_error"}

        # Get auth token from parameter or fall back to default
        auth_token = None
        if authorization and authorization.startswith("Bearer ") and authorization != "Bearer undefined" and authorization != "Bearer null":
            auth_token = authorization
            logger.info("Using provided authorization token")
        else:
            auth_token = f"Bearer {DEV_JWT_TOKEN}"
            logger.info(
                f"Using DEV_JWT_TOKEN for authorization: {auth_token[:20]}...")

        # Build query parameters - only include if provided
        params = {}
        if status:
            params["status"] = status
        if priority:
            params["priority"] = priority
        if page is not None:
            params["page"] = page
        if page_size is not None:
            params["page_size"] = page_size
        if user_id:  # Only include user_id if explicitly provided
            params["user_id"] = user_id

        # Define the get function for try_backend_urls
        async def get_func(client, url, **kwargs):
            return await client.get(url, **kwargs)

        # Determine endpoint based on item type
        endpoint = "/api/todo-lists" if item_type == "todos" else "/api/habits"

        # Use the enhanced try_backend_urls function
        result = await try_backend_urls(
            get_func,
            endpoint,
            headers={
                "Content-Type": "application/json",
                "Authorization": auth_token
            },
            params=params
        )

        # Check if the result is an error
        if result.get("status") == "error":
            await ctx.error(result.get("error", f"Unknown error fetching {item_type}"))
        else:
            await ctx.info(f"Successfully retrieved {item_type} data from backend")

        return result

    except Exception as e:
        error_msg = f"Error in get_items: {str(e)}"
        logger.error(error_msg)
        await ctx.error(error_msg)
        return {"status": "error", "error": error_msg, "type": "general_error"}


@mcp.tool("todos.create")
async def create_todo(
    ctx: Context,
    title: str,
    description: Optional[str] = None,
    due_date: Optional[str] = None,
    priority: Optional[str] = None,
    authorization: Optional[str] = None,
    user_id: Optional[str] = None,
    list_id: Optional[str] = None
) -> Dict[str, Any]:
    """Create a new todo item.

    Args:
        title: Title of the todo item
        description: Optional description
        due_date: Optional due date (format: YYYY-MM-DD)
        priority: Optional priority (high, medium, low)
        authorization: Optional authorization token (Bearer token)
        user_id: Optional user ID (will be extracted from token if not provided)
        list_id: Optional list ID for the todo

    Returns:
        The created todo item
    """
    try:
        # Get auth token from parameter or fall back to default
        auth_token = None
        if authorization and authorization.startswith("Bearer ") and authorization != "Bearer undefined" and authorization != "Bearer null":
            auth_token = authorization
            logger.info("Using provided authorization token")
        else:
            auth_token = f"Bearer {DEV_JWT_TOKEN}"
            logger.info(
                f"Using DEV_JWT_TOKEN for authorization: {auth_token[:20]}...")

        # Prepare todo data matching the Go backend's expected format
        todo_data = {
            "title": title,
            "description": description or "",
            "priority": priority or "medium",
            "status": "pending",
            "is_completed": False,
            "is_recurring": False,
            "due_date": due_date,
            "reminder_time": None,
            "recurrence_pattern": {},
            "tags": {},
            "checklist": {"items": []},
            "linked_task_id": None,
            "linked_calendar_event_id": None,
            "user_id": user_id,  # This will be required
            "list_id": list_id   # This will be required
        }

        async def post_func(client, url, **kwargs):
            return await client.post(url, **kwargs)

        return await try_backend_urls(
            post_func,
            "/api/todos",
            headers={
                "Content-Type": "application/json",
                "Authorization": auth_token
            },
            json=todo_data
        )
    except Exception as e:
        error_msg = f"Error creating todo: {str(e)}"
        logger.error(error_msg)
        await ctx.error(error_msg)
        return {"status": "error", "error": error_msg, "type": "api_error"}


@mcp.tool("habits.create")
async def create_habit(
    ctx: Context,
    title: str,
    description: Optional[str] = None,
    start_day: Optional[str] = None,
    end_day: Optional[str] = None,
    authorization: Optional[str] = None,
    user_id: Optional[str] = None
) -> Dict[str, Any]:
    """Create a new habit.

    Args:
        title: Title of the habit
        description: Optional description
        start_day: Optional start date (format: YYYY-MM-DD)
        end_day: Optional end date (format: YYYY-MM-DD) 
        authorization: Optional authorization token (Bearer token)
        user_id: Optional user ID (will be extracted from token if not provided)

    Returns:
        The created habit
    """
    try:
        # Get auth token from parameter or fall back to default
        auth_token = None
        if authorization and authorization.startswith("Bearer ") and authorization != "Bearer undefined" and authorization != "Bearer null":
            auth_token = authorization
            logger.info("Using provided authorization token")
        else:
            auth_token = f"Bearer {DEV_JWT_TOKEN}"
            logger.info(
                f"Using DEV_JWT_TOKEN for authorization: {auth_token[:20]}...")

        # Prepare habit data matching the Go backend's expected format
        habit_data = {
            "title": title,
            "description": description or "",
            "start_day": start_day,
            "end_day": end_day,
            "user_id": user_id
        }

        async def post_func(client, url, **kwargs):
            return await client.post(url, **kwargs)

        return await try_backend_urls(
            post_func,
            "/api/habits",
            headers={
                "Content-Type": "application/json",
                "Authorization": auth_token
            },
            json=habit_data
        )
    except Exception as e:
        error_msg = f"Error creating habit: {str(e)}"
        logger.error(error_msg)
        await ctx.error(error_msg)
        return {"status": "error", "error": error_msg, "type": "api_error"}


@mcp.tool("calendar.getEvents")
async def get_calendar_events(
    ctx: Context,
    start_date: str,
    end_date: Optional[str] = None,
    authorization: Optional[str] = None
) -> Dict[str, Any]:
    """Get calendar events for a specific date or date range.

    Args:
        start_date: Start date to fetch events from (format: YYYY-MM-DD)
        end_date: Optional end date to fetch events to (format: YYYY-MM-DD)
        authorization: Optional authorization token (Bearer token)

    Returns:
        List of calendar events in the specified date range
    """
    try:
        # Get auth token from parameter or fall back to default
        auth_token = None
        if authorization and authorization.startswith("Bearer ") and authorization != "Bearer undefined" and authorization != "Bearer null":
            auth_token = authorization
            logger.info("Using provided authorization token")
        else:
            auth_token = f"Bearer {DEV_JWT_TOKEN}"
            logger.info(
                f"Using DEV_JWT_TOKEN for authorization: {auth_token[:20]}...")

        # Format dates as RFC3339/ISO format with timezone
        # If only a date is provided (YYYY-MM-DD), convert to start/end of day with UTC timezone
        start_time = start_date
        if len(start_date) <= 10:  # It's just a date without time
            # Beginning of the day in UTC
            start_time = f"{start_date}T00:00:00Z"

        end_time = end_date
        if end_date:
            if len(end_date) <= 10:  # It's just a date without time
                end_time = f"{end_date}T23:59:59Z"  # End of the day in UTC
        elif start_date:
            # If no end_date is provided, default to end of the start_date
            end_date_val = start_date
            end_time = f"{end_date_val}T23:59:59Z"  # End of the start day

        # Build query parameters with correct names: start_time/end_time
        params = {"start_time": start_time}
        if end_time:
            params["end_time"] = end_time

        async def get_func(client, url, **kwargs):
            return await client.get(url, **kwargs)

        return await try_backend_urls(
            get_func,
            "/api/calendar/events",
            headers={
                "Content-Type": "application/json",
                "Authorization": auth_token
            },
            params=params
        )
    except Exception as e:
        error_msg = f"Error getting calendar events: {str(e)}"
        logger.error(error_msg)
        await ctx.error(error_msg)
        return {"status": "error", "error": error_msg, "type": "api_error"}


@mcp.tool("calendar.createEvent")
async def create_calendar_event(
    ctx: Context,
    title: str,
    start_time: str,
    end_time: str,
    description: Optional[str] = None,
    location: Optional[str] = None,
    is_all_day: Optional[bool] = False,
    event_type: Optional[str] = "Meeting",
    color: Optional[str] = "#3b82f6",
    transparency: Optional[str] = "opaque",
    authorization: Optional[str] = None,
    check_conflicts: Optional[bool] = True
) -> Dict[str, Any]:
    """Create a new calendar event with conflict checking.

    Args:
        title: Title of the event
        start_time: Start time of the event (ISO format)
        end_time: End time of the event (ISO format)
        description: Optional description
        location: Optional location of the event
        is_all_day: Whether the event is an all-day event
        event_type: Optional event type (None, Task, Meeting, Todo, Holiday, Reminder)
        color: Optional event color (hex format)
        transparency: Optional transparency ('opaque' or 'transparent')
        authorization: Optional authorization token (Bearer token)
        check_conflicts: Whether to check for scheduling conflicts

    Returns:
        The created calendar event with insights about conflicts
    """
    try:
        # Get auth token from parameter or fall back to default
        auth_token = None
        if authorization and authorization.startswith("Bearer ") and authorization != "Bearer undefined" and authorization != "Bearer null":
            auth_token = authorization
            logger.info("Using provided authorization token")
        else:
            auth_token = f"Bearer {DEV_JWT_TOKEN}"
            logger.info(
                f"Using DEV_JWT_TOKEN for authorization: {auth_token[:20]}...")

        # Ensure start_time and end_time have RFC3339 format
        if not any(s in start_time for s in ["Z", "+", "-"]) or "-" not in start_time[10:]:
            start_time = f"{start_time}+03:00"

        if not any(s in end_time for s in ["Z", "+", "-"]) or "-" not in end_time[10:]:
            end_time = f"{end_time}+03:00"

        # Extract date from start_time for conflict checking
        event_date = start_time.split(
            "T")[0] if "T" in start_time else start_time

        # Normalize event_type to a valid value (must match the Go backend's EventType enum)
        valid_event_types = ["None", "Task",
                             "Meeting", "Todo", "Holiday", "Reminder"]
        if not event_type or event_type not in valid_event_types:
            event_type = "Meeting"  # Default to Meeting

        # Check for conflicts if requested
        insights = {}
        if check_conflicts:
            await ctx.info(f"Checking for conflicts on {event_date}")

            # Get events for the same day
            async def get_func(client, url, **kwargs):
                return await client.get(url, **kwargs)

            # Format dates as RFC3339/ISO format
            # Beginning of the day in UTC
            conflict_start = f"{event_date}T00:00:00Z"
            conflict_end = f"{event_date}T23:59:59Z"    # End of the day in UTC

            events_result = await try_backend_urls(
                get_func,
                "/api/calendar/events",
                headers={
                    "Content-Type": "application/json",
                    "Authorization": auth_token
                },
                params={"start_time": conflict_start, "end_time": conflict_end}
            )

            # Process events to check for conflicts
            conflicts = []
            events = []
            if events_result.get("status") != "error" and isinstance(events_result, dict):
                events = events_result.get("events", [])
                if not isinstance(events, list):
                    events = []

                for event in events:
                    event_start = event.get("start_time", "")
                    event_end = event.get("end_time", "")
                    event_title = event.get("title", "Untitled Event")

                    # Check if there's an overlap
                    if ((event_start <= start_time <= event_end) or
                        (event_start <= end_time <= event_end) or
                            (start_time <= event_start and end_time >= event_end)):
                        conflicts.append({
                            "title": event_title,
                            "start_time": event_start,
                            "end_time": event_end
                        })

            # Add insights about the schedule
            insights = {
                "total_events_on_day": len(events),
                "conflicts": conflicts,
                "has_conflicts": len(conflicts) > 0,
                "conflict_count": len(conflicts)
            }

            if insights["has_conflicts"]:
                await ctx.warning(f"Found {len(conflicts)} scheduling conflicts for the requested time")
            else:
                await ctx.info("No scheduling conflicts found")

        # Prepare event data matching the exact expected format by the Go backend
        event_data = {
            "title": title,
            "description": description or "",
            "event_type": event_type,
            "start_time": start_time,
            "end_time": end_time,
            "is_all_day": is_all_day,
            "location": location or "",
            "color": color,
            "transparency": transparency
        }

        await ctx.info(f"Creating calendar event with data: {json.dumps(event_data, default=str)[:200]}...")

        async def post_func(client, url, **kwargs):
            return await client.post(url, **kwargs)

        # Create the event
        result = await try_backend_urls(
            post_func,
            "/api/calendar/events",
            headers={
                "Content-Type": "application/json",
                "Authorization": auth_token
            },
            json=event_data
        )

        # Add insights to the result
        if isinstance(result, dict):
            result["insights"] = insights

        return result
    except Exception as e:
        error_msg = f"Error creating calendar event: {str(e)}"
        logger.error(error_msg)
        await ctx.error(error_msg)
        return {"status": "error", "error": error_msg, "type": "api_error"}


@mcp.tool("todos.smartUpdate")
async def smart_update_todo_handler(
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
    return await smart_update_todo(ctx, edit_request, authorization, user_id)


@app.get("/api-test/jwt-check")
async def test_jwt():
    """Endpoint to test JWT token validity."""
    try:
        # Test JWT token
        auth_token = f"Bearer {DEV_JWT_TOKEN}"

        test_results = {
            "jwt_token": DEV_JWT_TOKEN[:20] + "...",
            "backend_urls": GO_BACKEND_URLS,
            "current_backend_url": GO_BACKEND_URL,
            "endpoints_tested": [],
        }

        # Try to access the todos endpoint with this token
        for base_url in GO_BACKEND_URLS:
            full_url = f"{base_url}/api/todo-lists"
            try:
                async with httpx.AsyncClient(timeout=5.0) as client:
                    headers = {
                        "Content-Type": "application/json",
                        "Authorization": auth_token
                    }
                    response = await client.get(full_url, headers=headers)

                    test_results["endpoints_tested"].append({
                        "url": full_url,
                        "status_code": response.status_code,
                        "response": response.text,
                        "success": response.status_code < 400
                    })
            except Exception as e:
                test_results["endpoints_tested"].append({
                    "url": full_url,
                    "error": str(e),
                    "success": False
                })

        # Try to login to get a new token
        login_url = f"{GO_BACKEND_URL}/api/users/login"
        try:
            login_data = {
                "email": "admin@example.com",
                "password": "password123"
            }
            async with httpx.AsyncClient(timeout=5.0) as client:
                response = await client.post(
                    login_url,
                    json=login_data,
                    headers={"Content-Type": "application/json"}
                )

                test_results["login_attempt"] = {
                    "url": login_url,
                    "status_code": response.status_code,
                    "response": response.text if response.status_code >= 400 else "Login response (contains sensitive data)",
                    "success": response.status_code < 400
                }

                if response.status_code < 400:
                    # Extract the token if login was successful
                    login_response = response.json()
                    if "token" in login_response:
                        test_results["new_token"] = login_response["token"][:20] + "..."
        except Exception as e:
            test_results["login_attempt"] = {
                "url": login_url,
                "error": str(e),
                "success": False
            }

        return test_results
    except Exception as e:
        return {"error": str(e)}


@app.post("/mcp/tools/refresh")
async def refresh_tools():
    """Endpoint to refresh tools cache."""
    try:
        await AICacheManager.invalidate_cache()
        return {"status": "success", "message": "Tools cache invalidated"}
    except Exception as e:
        logger.error(f"Error refreshing tools cache: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


@mcp.tool("notes.get")
async def get_notes(
    ctx: Context,
    page: int = 1,
    authorization: Optional[str] = None
) -> Dict[str, Any]:
    """Get notes using GraphQL matching frontend query structure.

    Args:
        page: Page number for pagination (starts at 1)
        authorization: Optional authorization token (Bearer token)

    Returns:
        Dictionary containing notes and pagination info matching frontend types
    """
    try:
        # Get auth token from parameter or fall back to default
        auth_token = None
        if authorization and authorization.startswith("Bearer ") and authorization != "Bearer undefined" and authorization != "Bearer null":
            auth_token = authorization
            logger.info("Using provided authorization token")
        else:
            auth_token = f"Bearer {DEV_JWT_TOKEN}"
            logger.info(
                f"Using DEV_JWT_TOKEN for authorization: {auth_token[:20]}...")

        # GraphQL query matching frontend GET_NOTES query exactly
        graphql_query = """
        query GetNotes($page: Int!) {
          notePages(page: $page) {
            success
            message
            data {
              id
              title
              content
              tags
              favorited
              createdAt
              updatedAt
            }
            pageInfo {
              totalPages
              totalItems
              currentPage
            }
          }
        }
        """

        # Prepare GraphQL request payload
        graphql_payload = {
            "query": graphql_query,
            "variables": {
                "page": page
            }
        }

        # Log the request
        await ctx.info(f"Sending GraphQL request with variables: {json.dumps(graphql_payload['variables'])}")

        # Execute GraphQL query
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f"{settings.NOTES_SERVER_URL}/graphql",
                headers={
                    "Content-Type": "application/json",
                    "Authorization": auth_token
                },
                json=graphql_payload
            )

            result = response.json()

        # Check for GraphQL errors
        if "errors" in result:
            error_msg = f"GraphQL errors: {json.dumps(result['errors'])}"
            logger.error(error_msg)
            await ctx.error(error_msg)
            return {"status": "error", "error": error_msg, "type": "graphql_error", "details": result["errors"]}

        # Extract and return just the notePages data matching frontend types
        if "data" in result and "notePages" in result["data"]:
            return result["data"]["notePages"]
        else:
            return result

    except Exception as e:
        error_msg = f"Error retrieving notes via GraphQL: {str(e)}"
        logger.error(error_msg)
        await ctx.error(error_msg)
        return {"status": "error", "error": error_msg, "type": "api_error"}


@mcp.tool("notes.create")
async def create_note(
    ctx: Context,
    title: str,
    content: str,
    tags: Optional[List[str]] = None,
    favorited: Optional[bool] = False,
    authorization: Optional[str] = None
) -> Dict[str, Any]:
    """Create a new note using GraphQL matching frontend mutation structure.

    Args:
        title: Title of the note
        content: Content of the note (markdown supported)
        tags: Optional list of tags to associate with the note
        favorited: Optional boolean indicating if the note should be favorited
        authorization: Optional authorization token (Bearer token)

    Returns:
        Dictionary containing the created note data
    """
    try:
        # Get auth token from parameter or fall back to default
        auth_token = None
        if authorization and authorization.startswith("Bearer ") and authorization != "Bearer undefined" and authorization != "Bearer null":
            auth_token = authorization
            logger.info("Using provided authorization token")
        else:
            auth_token = f"Bearer {DEV_JWT_TOKEN}"
            logger.info(
                f"Using DEV_JWT_TOKEN for authorization: {auth_token[:20]}...")

        # GraphQL mutation matching frontend CREATE_NOTE mutation exactly
        graphql_mutation = """
        mutation CreateNote($input: NotePageInput!) {
          createNotePage(input: $input) {
            success
            message
            data {
              id
              title
              content
              tags
              favorited
              createdAt
              updatedAt
            }
          }
        }
        """

        # Prepare input data matching the frontend types
        note_input = {
            "title": title,
            "content": content,
            "tags": tags or [],
            "favorited": favorited
        }

        # Prepare GraphQL request payload
        graphql_payload = {
            "query": graphql_mutation,
            "variables": {
                "input": note_input
            }
        }

        # Log the request
        await ctx.info(f"Sending GraphQL create note request: {title}")

        # Execute GraphQL mutation
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f"{settings.NOTES_SERVER_URL}/graphql",
                headers={
                    "Content-Type": "application/json",
                    "Authorization": auth_token
                },
                json=graphql_payload
            )

            result = response.json()

        # Check for GraphQL errors
        if "errors" in result:
            error_msg = f"GraphQL errors: {json.dumps(result['errors'])}"
            logger.error(error_msg)
            await ctx.error(error_msg)
            return {"status": "error", "error": error_msg, "type": "graphql_error", "details": result["errors"]}

        # Extract and return just the createNotePage data
        if "data" in result and "createNotePage" in result["data"]:
            return result["data"]["createNotePage"]
        else:
            return result

    except Exception as e:
        error_msg = f"Error creating note via GraphQL: {str(e)}"
        logger.error(error_msg)
        await ctx.error(error_msg)
        return {"status": "error", "error": error_msg, "type": "api_error"}


@mcp.tool("notes.rewriteInStyle")
async def rewrite_in_style(
    ctx: Context,
    text: str,
    user_id: Optional[str] = None,
    authorization: Optional[str] = None
) -> Dict[str, Any]:
    """Rewrite text in user's personal style based on their notes and journals.

    Args:
        text: The text to rewrite
        user_id: Optional user ID to fetch their writing style
        authorization: Optional authorization token (Bearer token)

    Returns:
        Dictionary containing the rewritten text
    """
    try:
        # Get auth token from parameter or fall back to default
        auth_token = None
        if authorization and authorization.startswith("Bearer ") and authorization != "Bearer undefined" and authorization != "Bearer null":
            auth_token = authorization
            logger.info("Using provided authorization token")
        else:
            auth_token = f"Bearer {DEV_JWT_TOKEN}"
            logger.info(
                f"Using DEV_JWT_TOKEN for authorization: {auth_token[:20]}...")

        # Enhanced GraphQL query to fetch user's notes for style analysis
        graphql_query = """
        query GetUserNotes($page: Int!) {
          notePages(page: $page) {
            data {
              content
              updatedAt
              title
            }
          }
        }
        """

        # Fetch user's notes for style analysis
        user_notes = []
        page = 1

        while True:
            graphql_payload = {
                "query": graphql_query,
                "variables": {"page": page}
            }

            async with httpx.AsyncClient() as client:
                response = await client.post(
                    f"{settings.NOTES_SERVER_URL}/graphql",
                    headers={
                        "Content-Type": "application/json",
                        "Authorization": auth_token
                    },
                    json=graphql_payload
                )

                result = response.json()

                if "data" in result and "notePages" in result["data"]:
                    notes_data = result["data"]["notePages"]["data"]
                    if not notes_data:
                        break

                    # Filter out empty or very short notes
                    valid_notes = [
                        {
                            "content": note["content"],
                            "updatedAt": note["updatedAt"],
                            "title": note["title"]
                        }
                        for note in notes_data
                        # Only use notes with substantial content
                        if note["content"] and len(note["content"]) > 50
                    ]
                    user_notes.extend(valid_notes)
                    page += 1

                    # Stop after collecting enough notes for analysis
                    if len(user_notes) >= 10:  # We only need 10 notes for style analysis
                        break
                else:
                    break

        if not user_notes:
            await ctx.warning("No valid notes found for style analysis. Using default style.")
            return {
                "status": "success",
                "content": {
                    "original_text": text,
                    "rewritten_text": text,  # Return original text if no style samples
                    "style_samples_used": 0
                }
            }

        # Sort notes by recency
        user_notes.sort(key=lambda x: x["updatedAt"], reverse=True)

        # Take the most recent notes with substantial content
        selected_notes = user_notes[:10]

        # Combine notes content for style analysis, including titles for better context
        style_samples = []
        for note in selected_notes:
            # Add title as a heading
            if note["title"]:
                style_samples.append(f"# {note['title']}")
            # Add content
            if note["content"]:
                style_samples.append(note["content"])

        style_context = "\n\n".join(style_samples)

        # Enhanced prompt for better style matching
        llm_service = LLMService()
        prompt = f"""You are tasked to only list the style context

{style_context}

Based on these examples, please rewrite the following text to match the user's:
1. Writing tone and voice
2. Typical sentence structure and length
3. Word choice and vocabulary preferences
4. Formatting patterns and emphasis styles
5. Any unique expressions or phrases they commonly use

Text to rewrite:
{text}

Return ONLY the rewritten text, without any additional formatting or metadata."""

        response = await llm_service.generate_response(
            prompt=prompt,
            context={
                "task": "style_rewrite",
                "original_text": text,
                "style_samples_count": len(selected_notes)
            },
            user_id=user_id
        )

        # Parse the response
        if isinstance(response, dict):
            if "text" in response:
                rewritten_text = response["text"]
            elif "content" in response:
                rewritten_text = response["content"]
            else:
                rewritten_text = str(response)
        else:
            rewritten_text = str(response)

        # Clean up the response if it's a string representation of a dict
        if isinstance(rewritten_text, str):
            try:
                # Try to parse if it looks like a dict string
                if rewritten_text.startswith("{") and rewritten_text.endswith("}"):
                    parsed = json.loads(rewritten_text.replace("'", '"'))
                    if isinstance(parsed, dict) and "text" in parsed:
                        rewritten_text = parsed["text"]
            except:
                # If parsing fails, use the string as is
                pass

        # Remove any HTML tags if present
        rewritten_text = rewritten_text.replace("<p>", "").replace("</p>", "")

        return {
            "status": "success",
            "content": {
                "original_text": text,
                "rewritten_text": rewritten_text,
                "style_samples_used": len(selected_notes),
                "total_notes_analyzed": len(user_notes)
            }
        }

    except Exception as e:
        error_msg = f"Error rewriting text in style: {str(e)}"
        logger.error(error_msg)
        await ctx.error(error_msg)
        return {"status": "error", "error": error_msg, "type": "api_error"}


@mcp.tool("todos.addChecklist")
async def add_todo_checklist(
    ctx: Context,
    todo_id: str,
    checklist_items: List[str],
    authorization: Optional[str] = None
) -> Dict[str, Any]:
    """Add checklist items to an existing todo.

    Args:
        todo_id: ID of the todo to update
        checklist_items: List of checklist item descriptions to add
        authorization: Optional authorization token (Bearer token)

    Returns:
        The updated todo item with the new checklist
    """
    try:
        # Get auth token from parameter or fall back to default
        auth_token = None
        if authorization and authorization.startswith("Bearer ") and authorization != "Bearer undefined" and authorization != "Bearer null":
            auth_token = authorization
            logger.info("Using provided authorization token")
        else:
            auth_token = f"Bearer {DEV_JWT_TOKEN}"
            logger.info(
                f"Using DEV_JWT_TOKEN for authorization: {auth_token[:20]}...")

        # Prepare checklist data in the format expected by the backend
        checklist_data = {
            "checklist": {
                "items": [{"title": item, "completed": False} for item in checklist_items]
            }
        }

        async def put_func(client, url, **kwargs):
            return await client.put(url, **kwargs)

        return await try_backend_urls(
            put_func,
            f"/api/todos/{todo_id}",
            headers={
                "Content-Type": "application/json",
                "Authorization": auth_token
            },
            json=checklist_data
        )
    except Exception as e:
        error_msg = f"Error adding checklist items to todo: {str(e)}"
        logger.error(error_msg)
        await ctx.error(error_msg)
        return {"status": "error", "error": error_msg, "type": "api_error"}


def setup_mcp_server(app: Optional[FastAPI] = None):
    """Setup and return the MCP server instance"""
    # Log basic info
    logger.info(f"Setting up MCP server: {mcp.name}")

    # Invalidate cache on startup
    async def invalidate_cache_on_startup():
        try:
            await AICacheManager.invalidate_cache()
            logger.info("Invalidated AI cache on startup")
        except Exception as e:
            logger.error(f"Error invalidating cache on startup: {str(e)}")

    # Add startup event
    if app:
        app.add_event_handler("startup", invalidate_cache_on_startup)

    return mcp


async def run_server():
    """Run the MCP server with stdio transport."""
    try:
        logger.info("Starting MCP server with stdio transport")
        await mcp.run_stdio_async()
    except Exception as e:
        logger.error(f"Error running MCP server: {str(e)}", exc_info=True)
        sys.exit(1)


if __name__ == "__main__":
    # Run the server using stdio transport
    logger.info("Initializing MCP server")
    asyncio.run(run_server())
