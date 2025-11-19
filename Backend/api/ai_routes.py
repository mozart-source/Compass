from fastapi import APIRouter, HTTPException, status, Query, File, UploadFile, Request, BackgroundTasks, Depends, Cookie, Header
from typing import Dict, List, Optional, Any, Union, AsyncIterator
from datetime import datetime
from fastapi.responses import StreamingResponse, JSONResponse
from pydantic import BaseModel, Field
import logging
import os
import shutil
import json
from pathlib import Path
import uuid
import asyncio
import jwt

from ai_services.llm.llm_service import LLMService
from orchestration.ai_orchestrator import AIOrchestrator
from app.schemas.message_schemas import UserMessage, AssistantMessage, ConversationHistory
from core.config import settings
from core.mcp_state import get_mcp_client

# Set up logger
logger = logging.getLogger(__name__)

# Request/Response Models


class PreviousMessage(BaseModel):
    sender: str
    text: str


class AIRequest(BaseModel):
    prompt: str
    context: Optional[Dict] = None
    domain: Optional[str] = None
    model_parameters: Optional[Dict] = None
    previous_messages: Optional[List[PreviousMessage]] = None
    session_id: Optional[str] = None  # Add session ID field


class AIResponse(BaseModel):
    response: str
    tool_used: Optional[str] = None
    tool_args: Optional[Dict[str, Any]] = None
    tool_success: Optional[bool] = None
    description: Optional[str] = None
    rag_used: bool = False
    cached: bool = False
    confidence: float = 0.0
    error: Optional[bool] = None
    error_message: Optional[str] = None
    session_id: Optional[str] = None  # Add session ID to response


class FeedbackRequest(BaseModel):
    feedback_score: float = Field(..., ge=0, le=1)
    feedback_text: Optional[str] = None


class ProcessPDFResponse(BaseModel):
    status: str
    message: str
    processed_files: Optional[List[Dict[str, Any]]] = None
    error: Optional[str] = None


class RewriteRequest(BaseModel):
    text: str
    user_id: Optional[str] = None


router = APIRouter(prefix="/ai", tags=["AI Services"])

# Initialize services
llm_service = LLMService()

# Store orchestrators by session ID for persistent conversations
orchestrator_instances: Dict[str, AIOrchestrator] = {}


def get_or_create_orchestrator(session_id: str) -> AIOrchestrator:
    """Get an existing orchestrator or create a new one for the session."""
    if session_id not in orchestrator_instances:
        logger.info(f"Creating new orchestrator for session {session_id}")
        orchestrator_instances[session_id] = AIOrchestrator()
    return orchestrator_instances[session_id]


@router.get("/rag/stats/{domain}")
async def get_rag_stats(
    domain: str,
):
    """Get RAG statistics for a specific domain through MCP."""
    try:
        mcp_client = get_mcp_client()
        if not mcp_client:
            raise HTTPException(
                status_code=503, detail="MCP client not initialized")

        result = await mcp_client.call_tool("rag.stats", {
            "domain": domain
        })

        # Check if the tool call was successful
        if result.get("status") != "success":
            error_message = result.get(
                "error", "Unknown error calling RAG stats tool")
            logger.error(f"RAG stats tool error: {error_message}")
            raise HTTPException(status_code=500, detail=error_message)

        # Extract content from the result
        content = result.get("content", {})
        if isinstance(content, str):
            try:
                content = json.loads(content)
            except json.JSONDecodeError:
                logger.warning(
                    f"Could not parse RAG stats content as JSON: {content}")

        return content
    except HTTPException:
        raise
    except Exception as e:
        logger.error(
            f"Error getting RAG stats through MCP: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/rag/update/{domain}")
async def update_rag_knowledge(
    domain: str,
    content: Dict,
):
    """Update the RAG knowledge base for a domain through MCP."""
    try:
        mcp_client = get_mcp_client()
        if not mcp_client:
            raise HTTPException(
                status_code=503, detail="MCP client not initialized")

        result = await mcp_client.call_tool("rag.update", {
            "domain": domain,
            "content": content
        })

        # Check if the tool call was successful
        if result.get("status") != "success":
            error_message = result.get(
                "error", "Unknown error updating RAG knowledge")
            logger.error(f"RAG update tool error: {error_message}")
            raise HTTPException(status_code=500, detail=error_message)

        # Extract content from the result
        content = result.get("content", {})
        if isinstance(content, str):
            try:
                content = json.loads(content)
            except json.JSONDecodeError:
                logger.warning(
                    f"Could not parse RAG update content as JSON: {content}")

        return content
    except HTTPException:
        raise
    except Exception as e:
        logger.error(
            f"Error updating RAG knowledge through MCP: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))


@router.get("/model/info")
async def get_model_info():
    """Get information about the AI model configuration through MCP."""
    try:
        mcp_client = get_mcp_client()
        if not mcp_client:
            raise HTTPException(
                status_code=503, detail="MCP client not initialized")

        result = await mcp_client.call_tool("ai.model.info", {})

        # Check if the tool call was successful
        if result.get("status") != "success":
            error_message = result.get(
                "error", "Unknown error getting model info")
            logger.error(f"Model info tool error: {error_message}")
            raise HTTPException(status_code=500, detail=error_message)

        # Extract content from the result
        content = result.get("content", {})
        if isinstance(content, str):
            try:
                content = json.loads(content)
            except json.JSONDecodeError:
                logger.warning(
                    f"Could not parse model info content as JSON: {content}")

        return content
    except HTTPException:
        raise
    except Exception as e:
        logger.error(
            f"Error getting model info through MCP: {str(e)}", exc_info=True)
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/clear-session")
async def clear_session(
    request: Request
):
    """Clear conversation history for a session."""
    try:
        # Parse request body
        data = await request.json()
        session_id = data.get("session_id")

        if not session_id:
            raise HTTPException(
                status_code=400, detail="Missing session_id in request body")

        if session_id in orchestrator_instances:
            orchestrator = orchestrator_instances[session_id]
            user_id = int(hash(session_id) % 100000)

            # Use the mongo_client directly to clear the conversation
            try:
                # Get conversation by session ID
                conversation = orchestrator.mongo_client.get_conversation_by_session(
                    session_id)
                if conversation and conversation.id:
                    # Clear messages for the conversation
                    orchestrator.mongo_client.conversation_repo.update(
                        conversation.id,
                        {"messages": []}
                    )
                    logger.info(
                        f"Cleared conversation history for session {session_id}")
                    return {"status": "success", "message": f"Session {session_id} cleared"}
                else:
                    logger.warning(
                        f"No conversation found for session {session_id}")
                    return {"status": "not_found", "message": f"No conversation found for session {session_id}"}
            except Exception as e:
                logger.error(f"Error clearing conversation: {str(e)}")
                raise HTTPException(
                    status_code=500, detail=f"Error clearing conversation: {str(e)}")

        logger.warning(f"Session {session_id} not found for clearing")
        return {"status": "not_found", "message": f"Session {session_id} not found"}
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error clearing session: {str(e)}")
        raise HTTPException(
            status_code=500, detail=f"Error clearing session: {str(e)}")


@router.post("/rag/knowledge-base/process", response_model=Dict[str, Any], status_code=202)
async def process_knowledge_base(
    request: Request,
    background_tasks: BackgroundTasks,
    domain: Optional[str] = None,
):
    """Process knowledge base files through MCP."""
    try:
        mcp_client = get_mcp_client()
        if not mcp_client:
            raise HTTPException(
                status_code=503, detail="MCP client not initialized")

        result = await mcp_client.call_tool("rag.knowledge-base.process", {
            "domain": domain
        })
        return result
    except Exception as e:
        logger.error(f"Error processing knowledge base through MCP: {str(e)}")
        raise HTTPException(
            status_code=500, detail=f"Error processing knowledge base: {str(e)}")


@router.post("/knowledge-base/upload", response_model=ProcessPDFResponse)
async def upload_pdf_to_knowledge_base(
    file: UploadFile = File(...),
    domain: Optional[str] = None,
) -> ProcessPDFResponse:
    """Upload a PDF file to the knowledge base through MCP."""
    try:
        # Read file content
        content = await file.read()

        # Send file to MCP
        mcp_client = get_mcp_client()
        if not mcp_client:
            raise HTTPException(
                status_code=503, detail="MCP client not initialized")

        result = await mcp_client.call_tool("knowledge-base.upload", {
            "filename": file.filename,
            "content": content,
            "domain": domain
        })

        # Check if the tool call was successful
        if result.get("status") != "success":
            error_message = result.get("error", "Unknown error uploading PDF")
            logger.error(f"PDF upload tool error: {error_message}")
            raise HTTPException(status_code=500, detail=error_message)

        # Ensure content is properly parsed from the result
        processed_files = []
        content_data = result.get("content", {})

        # Handle string content that might be JSON
        if isinstance(content_data, str):
            try:
                content_data = json.loads(content_data)
            except json.JSONDecodeError:
                logger.error("Could not parse content as JSON")

        # Extract files list from the content
        if isinstance(content_data, dict) and "files" in content_data:
            processed_files = content_data["files"]
        elif isinstance(content_data, list):
            # If content is already a list, assume it's the files list
            processed_files = content_data

        return ProcessPDFResponse(
            status="success",
            message=f"PDF processed successfully: {file.filename}",
            processed_files=processed_files
        )
    except HTTPException:
        raise
    except Exception as e:
        logger.error(
            f"Error processing PDF through MCP: {str(e)}", exc_info=True)
        raise HTTPException(
            status_code=500,
            detail=f"Error processing PDF: {str(e)}"
        )


@router.post("/entity/create", response_model=AIResponse)
async def create_entity(
    request: AIRequest,
    session_id: Optional[str] = Cookie(None),
    authorization: Optional[str] = Header(None)
) -> AIResponse:
    """Create a new entity through MCP."""
    try:
        # Use session ID if provided
        active_session_id = request.session_id or session_id or str(
            uuid.uuid4())

        mcp_client = get_mcp_client()
        if not mcp_client:
            raise HTTPException(
                status_code=503, detail="MCP client not initialized")

        result = await mcp_client.call_tool("entity.create", {
            "prompt": request.prompt,
            "domain": request.domain or "default",
            "authorization": authorization
        })

        # Parse the response content properly
        response_content = {}
        if isinstance(result, dict) and "content" in result:
            content = result["content"]
            if isinstance(content, str):
                try:
                    response_content = json.loads(content)
                except json.JSONDecodeError:
                    response_content = {"response": content}
            elif isinstance(content, dict):
                response_content = content

        # Extract values with proper type handling
        response_text = response_content.get("response", "Entity created") if isinstance(
            response_content, dict) else "Entity created"
        description = response_content.get("description", "Create entity from description") if isinstance(
            response_content, dict) else "Create entity from description"
        rag_used = bool(response_content.get("rag_used", False)
                        ) if isinstance(response_content, dict) else False
        cached = bool(response_content.get("cached", False)) if isinstance(
            response_content, dict) else False
        confidence = float(response_content.get("confidence", 0.9)) if isinstance(
            response_content, dict) else 0.9
        error = bool(response_content.get("error", False)) if isinstance(
            response_content, dict) else False
        error_message = response_content.get(
            "error_message") if isinstance(response_content, dict) else None

        return AIResponse(
            response=response_text,
            description=description,
            rag_used=rag_used,
            cached=cached,
            confidence=confidence,
            error=error,
            error_message=error_message,
            session_id=active_session_id
        )
    except Exception as e:
        logger.error(f"Error creating entity through MCP: {str(e)}")
        return AIResponse(
            response=f"Error creating entity: {str(e)}",
            description="Create entity from description",
            rag_used=False,
            cached=False,
            confidence=0.0,
            error=True,
            error_message=str(e),
            session_id=request.session_id or session_id
        )


@router.post("/process/stream")
async def process_ai_request_stream(
    request: AIRequest,
    request_obj: Request,
    session_id: Optional[str] = Cookie(None),
    authorization: Optional[str] = Header(None)
) -> StreamingResponse:
    """Process an AI request and stream the response using Server-Sent Events (SSE)."""
    try:
        # Get or create session ID
        active_session_id = request.session_id or session_id or str(
            uuid.uuid4())

        logger.info(
            f"------- Received STREAMING request with session ID: {active_session_id} -------")
        logger.info(
            f"Streaming prompt: {request.prompt[:50]}{'...' if len(request.prompt) > 50 else ''}")

        # Extract user information from JWT token
        user_id = None
        organization_id = None
        if authorization and authorization.startswith("Bearer "):
            try:
                token = authorization.split(" ")[1]
                claims = jwt.decode(token, settings.jwt_secret_key, algorithms=[
                                    settings.jwt_algorithm])
                user_id = claims.get("user_id")
                organization_id = claims.get("org_id")
                logger.info(f"Extracted user_id: {user_id} from token")
            except Exception as e:
                logger.warning(f"Failed to decode JWT token: {str(e)}")

        # Get client information from the request object
        client_ip = request_obj.client.host if request_obj.client else None
        user_agent = request_obj.headers.get("user-agent")

        # Get or create orchestrator for this session
        orchestrator = get_or_create_orchestrator(active_session_id)

        # Map session ID to a numeric user ID for the orchestrator
        # Use a hash of the session ID to get a consistent integer
        orchestrator_user_id = int(hash(active_session_id) % 100000)
        logger.info(
            f"Processing streaming request for user_id: {orchestrator_user_id}")

        # Create the streaming response
        async def event_generator():
            """Generate SSE events from the LLM response."""
            try:
                token_count = 0
                # Stream the response from the orchestrator
                logger.info("Starting token stream from orchestrator")
                async for token in orchestrator.process_request_stream(
                    user_input=request.prompt,
                    user_id=orchestrator_user_id,
                    domain=request.domain or "default",
                    auth_token=authorization,  # Pass the authorization header
                    client_ip=client_ip,
                    user_agent=user_agent,
                    real_user_id=user_id,
                    organization_id=organization_id
                ):
                    # Format as SSE event
                    if token:
                        token_count += 1
                        if token_count % 10 == 0:
                            logger.debug(
                                f"Streamed {token_count} tokens so far")
                        yield f"data: {json.dumps(token)}\n\n"
                    # Add small delay to avoid overwhelming the client
                    await asyncio.sleep(0.01)

                logger.info(f"Stream complete - sent {token_count} tokens")
                # Send completion event
                yield f"data: [DONE]\n\n"
            except Exception as e:
                logger.error(
                    f"Error in streaming response: {str(e)}", exc_info=True)
                error_data = json.dumps({"error": str(e)})
                yield f"data: {error_data}\n\n"
                yield f"data: [DONE]\n\n"

        # Return streaming response
        logger.info("Initializing SSE streaming response")
        return StreamingResponse(
            event_generator(),
            media_type="text/event-stream",
            headers={
                "Cache-Control": "no-cache",
                "Connection": "keep-alive",
                "Content-Type": "text/event-stream",
                "X-Accel-Buffering": "no"
            }
        )
    except Exception as e:
        logger.error(
            f"Error setting up streaming AI request: {str(e)}", exc_info=True)
        # Return error as an event stream for consistent error handling

        async def error_generator():
            # Capture the error message from the outer scope
            error_message = str(e)
            error_data = json.dumps({"error": error_message})
            yield f"data: {error_data}\n\n"
            yield f"data: [DONE]\n\n"

        return StreamingResponse(
            error_generator(),
            media_type="text/event-stream",
            headers={
                "Cache-Control": "no-cache",
                "Connection": "keep-alive",
                "Content-Type": "text/event-stream"
            }
        )


@router.post("/rewrite-in-style")
async def rewrite_in_style(
    request: RewriteRequest,
    authorization: Optional[str] = Header(None)
) -> Dict[str, Any]:
    """Endpoint to rewrite text in user's personal style."""
    try:
        mcp_client = get_mcp_client()
        if not mcp_client:
            raise HTTPException(
                status_code=503, detail="MCP client not initialized")

        result = await mcp_client.call_tool("notes.rewriteInStyle", {
            "text": request.text,
            "user_id": request.user_id,
            "authorization": authorization
        })

        # Handle the MCP tool response which comes as a list of TextContent
        if hasattr(result, 'content') and isinstance(result.content, list):
            # Extract the text content from the first item
            text_content = result.content[0].text if result.content else None
            if text_content:
                try:
                    # Parse the JSON string into a dictionary
                    parsed_result = json.loads(text_content)
                    return parsed_result
                except json.JSONDecodeError:
                    # If parsing fails, return the raw text
                    return {"status": "success", "content": {"rewritten_text": text_content}}

        # If result is already a dictionary
        if isinstance(result, dict):
            return result

        # Fallback for unexpected response format
        return {"status": "error", "error": "Unexpected response format"}

    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error in rewrite-in-style endpoint: {str(e)}")
        raise HTTPException(status_code=500, detail=str(e))


# Add diagnostic endpoints
@router.get("/diagnostics")
async def ai_service_diagnostics():
    """Comprehensive AI service diagnostics endpoint."""
    try:
        diagnostics = {
            "timestamp": datetime.utcnow().isoformat(),
            "service_status": "running",
            "components": {}
        }

        # Check MCP Client
        try:
            from core.mcp_state import get_mcp_client
            mcp_client = get_mcp_client()
            if mcp_client:
                tools = await mcp_client.get_tools()
                diagnostics["components"]["mcp_client"] = {
                    "status": "connected",
                    "tools_count": len(tools),
                    "tools": [tool.get('name', 'Unknown') for tool in tools],
                    "sample_tools": [
                        {
                            "name": tool.get('name', 'Unknown'),
                            "description": tool.get('description', 'No description')[:100] + "..." if len(tool.get('description', '')) > 100 else tool.get('description', 'No description')
                        } for tool in tools[:5]
                    ]
                }

                # Test a simple MCP call
                try:
                    health_result = await mcp_client.call_tool("check.health")
                    diagnostics["components"]["mcp_health_test"] = {
                        "status": "success",
                        "result": health_result
                    }
                except Exception as health_error:
                    diagnostics["components"]["mcp_health_test"] = {
                        "status": "failed",
                        "error": str(health_error)
                    }
            else:
                diagnostics["components"]["mcp_client"] = {
                    "status": "not_connected",
                    "error": "MCP client not found in global state"
                }
        except Exception as mcp_error:
            diagnostics["components"]["mcp_client"] = {
                "status": "error",
                "error": str(mcp_error)
            }

        # Check LLM Service
        try:
            llm_info = await llm_service.get_model_info()
            diagnostics["components"]["llm_service"] = {
                "status": "connected",
                "model": llm_info.get("model", "unknown"),
                "capabilities": llm_info.get("capabilities", {})
            }
        except Exception as llm_error:
            diagnostics["components"]["llm_service"] = {
                "status": "error",
                "error": str(llm_error)
            }

        # Check Agent Orchestrator
        try:
            from ai_services.agents.orchestrator import AgentOrchestrator
            orchestrator = AgentOrchestrator()
            diagnostics["components"]["agent_orchestrator"] = {
                "status": "initialized",
                "entity_agents": list(orchestrator.entity_agents.keys()),
                "specialized_agents": list(orchestrator.specialized_agents.keys())
            }
        except Exception as orchestrator_error:
            diagnostics["components"]["agent_orchestrator"] = {
                "status": "error",
                "error": str(orchestrator_error)
            }

        # Overall status
        all_components_ok = all(
            comp.get("status") in ["connected", "initialized", "success"]
            for comp in diagnostics["components"].values()
        )
        diagnostics["overall_status"] = "healthy" if all_components_ok else "degraded"

        return diagnostics

    except Exception as e:
        logger.error(f"Error in AI diagnostics: {str(e)}")
        return {
            "timestamp": datetime.utcnow().isoformat(),
            "service_status": "error",
            "error": str(e)
        }
