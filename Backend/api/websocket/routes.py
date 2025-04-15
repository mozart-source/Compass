from fastapi import APIRouter, WebSocket, WebSocketDisconnect, Query, Depends, HTTPException, status
from typing import Optional, Dict, Any
import logging
import json
import asyncio
import traceback
from datetime import datetime

from api.websocket.dashboard_ws import dashboard_ws_manager
from core.config import settings
from ai_services.agents.orchestrator import AgentOrchestrator
from core.auth.jwt_handler import get_current_user, get_token_from_websocket
from core.mcp_state import get_mcp_client

logger = logging.getLogger(__name__)
router = APIRouter()

# Store active tasks to prevent garbage collection
active_tasks: Dict[str, asyncio.Task] = {}


@router.websocket("/ws")
async def dashboard_websocket(websocket: WebSocket):
    """
    WebSocket endpoint for dashboard real-time updates.
    Requires a valid JWT token as a query parameter.
    """
    try:
        token_data = await get_token_from_websocket(websocket)
        if not token_data:
            return

        user = get_current_user(token_data)
        user_id = user["user_id"]
        token = token_data["raw"]
    except Exception as auth_error:
        logger.error(
            f"WebSocket authentication failed: {auth_error}", exc_info=True)
        return

    try:
        logger.info(f"Token validated for user_id: {user_id}")

        # Accept the connection
        await dashboard_ws_manager.connect(websocket, user_id)
        logger.info(f"WebSocket connection accepted for user_id: {user_id}")

        # Verify that the connection is still alive after connect
        if websocket.client_state.name != "CONNECTED":
            logger.error(
                f"WebSocket connection failed immediately after accept for user_id: {user_id}")
            return

        # Send initial connection confirmation
        await websocket.send_json({
            "type": "connected",
            "timestamp": datetime.utcnow().isoformat(),
            "message": "Connected to dashboard updates"
        })

        # Send initial metrics immediately upon connection
        try:
            # Check if metrics are in memory cache first
            from data_layer.cache.dashboard_cache import dashboard_cache
            if hasattr(dashboard_cache, '_memory_cache') and user_id in dashboard_cache._memory_cache:
                metrics, _ = dashboard_cache._memory_cache[user_id]

                # Verify connection is still alive
                if websocket.client_state.name != "CONNECTED":
                    logger.warning(
                        f"WebSocket disconnected before sending initial metrics to user {user_id}")
                    return

                await websocket.send_json({
                    "type": "initial_metrics",
                    "data": metrics,
                    "timestamp": datetime.utcnow().isoformat()
                })
                logger.debug(
                    f"Sent initial metrics from memory cache to user {user_id}")
            else:
                # Fetch metrics if not in memory cache
                try:
                    metrics = await dashboard_cache.get_metrics(user_id, token)

                    # Verify connection is still alive
                    if websocket.client_state.name != "CONNECTED":
                        logger.warning(
                            f"WebSocket disconnected before sending fetched metrics to user {user_id}")
                        return

                    if metrics:
                        await websocket.send_json({
                            "type": "initial_metrics",
                            "data": metrics,
                            "timestamp": datetime.utcnow().isoformat()
                        })
                        logger.debug(f"Sent initial metrics to user {user_id}")
                    else:
                        # Send empty metrics structure if fetch fails
                        await websocket.send_json({
                            "type": "initial_metrics",
                            "data": {
                                "habits": None,
                                "calendar": None,
                                "focus": None,
                                "mood": None,
                                "ai_usage": None,
                                "system_metrics": None,
                                "goals": None,
                                "tasks": None,
                                "todos": None,
                                "user": None,
                                "notes": None,
                                "journals": None,
                                "cost": None,
                                "daily_timeline": None,
                                "timestamp": datetime.utcnow().isoformat()
                            },
                            "timestamp": datetime.utcnow().isoformat(),
                            "error": "Failed to fetch metrics from services"
                        })
                        logger.warning(
                            f"Sent empty metrics structure to user {user_id} due to fetch failure")
                except Exception as fetch_error:
                    logger.error(
                        f"Error fetching initial metrics: {str(fetch_error)}")

                    # Verify connection is still alive
                    if websocket.client_state.name != "CONNECTED":
                        logger.warning(
                            f"WebSocket disconnected after metrics fetch error for user {user_id}")
                        return

                    # Send error notification to client
                    await websocket.send_json({
                        "type": "error",
                        "message": "Failed to fetch dashboard metrics",
                        "timestamp": datetime.utcnow().isoformat()
                    })
        except Exception as e:
            logger.error(f"Error sending initial metrics: {str(e)}")
            # Don't disconnect on metrics error, just log it

        # Keep the connection alive and handle client messages
        while True:
            # Verify connection before waiting for message
            if websocket.client_state.name != "CONNECTED":
                logger.warning(
                    f"WebSocket connection lost for user {user_id} before receiving message")
                break

            # Wait for messages from the client
            try:
                data = await websocket.receive_text()
            except WebSocketDisconnect:
                logger.info(
                    f"WebSocket disconnected while waiting for message from user {user_id}")
                break
            except Exception as recv_error:
                logger.error(
                    f"Error receiving WebSocket message: {str(recv_error)}")
                break

            try:
                message = json.loads(data)
                message_type = message.get("type")
                message_data = message.get("data", {})

                # Handle different message types
                if message_type == "ping":
                    await websocket.send_json({"type": "pong", "timestamp": datetime.utcnow().isoformat()})
                elif message_type == "refresh":
                    # Client is requesting a refresh of dashboard data
                    logger.info(
                        f"Client requested dashboard refresh: {user_id}")
                    # Invalidate cache to force refresh on next API call
                    from data_layer.cache.dashboard_cache import dashboard_cache
                    await dashboard_cache.invalidate_cache(user_id)
                    await websocket.send_json({
                        "type": "refresh_initiated",
                        "timestamp": datetime.utcnow().isoformat()
                    })

                # AI Drag & Drop feature message handlers
                elif message_type == "ai_options_request":
                    # Use our new dedicated handler function
                    await handle_ai_options_request(websocket, message, user_id, token)

                elif message_type == "ai_process_request":
                    logger.info(
                        f"AI process request for option {message.get('option_id')} on {message.get('target_type')} {message.get('target_id')} from user {user_id}")

                    # Process the request asynchronously
                    task = asyncio.create_task(
                        process_ai_option(
                            websocket=websocket,
                            message_data=message,
                            user_id=user_id,
                            token=token
                        )
                    )
                    # Store the task to prevent it from being garbage collected
                    task_key = f"{user_id}_{message.get('target_id')}_{message.get('option_id')}"
                    active_tasks[task_key] = task

                elif message_type == "refresh_heatmap":
                    # Client is requesting a refresh of the habit heatmap specifically
                    logger.info(
                        f"Client requested habit heatmap refresh: {user_id}")
                    # Invalidate cache to force refresh on next API call
                    from data_layer.cache.dashboard_cache import dashboard_cache
                    await dashboard_cache.invalidate_cache(user_id)
                    await websocket.send_json({
                        "type": "heatmap_refresh_initiated",
                        "timestamp": datetime.utcnow().isoformat()
                    })
                    # Fetch fresh metrics after invalidation
                    metrics = await dashboard_cache.get_metrics(user_id, token)
                    if metrics and metrics.get("habit_heatmap"):
                        await websocket.send_json({
                            "type": "heatmap_data",
                            "data": {"habit_heatmap": metrics["habit_heatmap"]},
                            "timestamp": datetime.utcnow().isoformat()
                        })
                elif message_type == "refresh_focus":
                    # Client is requesting a refresh of the focus data specifically
                    logger.info(
                        f"Client requested focus data refresh: {user_id}")
                    # Invalidate cache to force refresh on next API call
                    from data_layer.cache.dashboard_cache import dashboard_cache
                    await dashboard_cache.invalidate_cache(user_id)
                    await websocket.send_json({
                        "type": "focus_refresh_initiated",
                        "timestamp": datetime.utcnow().isoformat()
                    })
                    # Fetch fresh metrics after invalidation - with high priority
                    try:
                        # Get focus metrics directly for faster response
                        from data_layer.repos.focus_repo import FocusSessionRepository, FocusSettingsRepository
                        focus_repo = FocusSessionRepository()
                        settings_repo = FocusSettingsRepository()

                        # Get focus stats with minimal latency
                        stats = focus_repo.get_stats(user_id)
                        user_settings = settings_repo.get_user_settings(
                            user_id)

                        # Convert stats to a dictionary we can modify safely
                        response_data = {}
                        if isinstance(stats, dict):
                            # If it's already a dict, make a copy
                            response_data = dict(stats)
                        else:
                            # Otherwise try to convert it
                            try:
                                response_data = dict(vars(stats))
                            except:
                                # Fallback to empty dict with key stats if available
                                response_data = {}
                                if hasattr(stats, "total_focus_seconds"):
                                    response_data["total_focus_seconds"] = stats.total_focus_seconds
                                if hasattr(stats, "streak"):
                                    response_data["streak"] = stats.streak
                                if hasattr(stats, "longest_streak"):
                                    response_data["longest_streak"] = stats.longest_streak
                                if hasattr(stats, "sessions"):
                                    response_data["sessions"] = stats.sessions

                        # Add settings data
                        response_data["daily_target_seconds"] = user_settings.daily_target_seconds

                        # Get daily breakdown - this returns a list of day objects, not an integer
                        breakdown = dashboard_cache._generate_daily_focus_breakdown(
                            user_id)
                        # Store it in a separate field so we don't overwrite any existing data
                        response_data = {**response_data,
                                         "focus_breakdown": breakdown}

                        # Send immediate response for better UX
                        await websocket.send_json({
                            "type": "focus_stats",
                            "data": response_data,
                            "timestamp": datetime.utcnow().isoformat()
                        })

                        # Then fetch complete metrics for cache update
                        metrics = await dashboard_cache.get_metrics(user_id, token)
                        if metrics and metrics.get("focus"):
                            await websocket.send_json({
                                "type": "focus_data",
                                "data": {"focus": metrics["focus"]},
                                "timestamp": datetime.utcnow().isoformat()
                            })
                    except Exception as e:
                        logger.error(f"Error fetching focus data: {e}")
                        # Fall back to getting all metrics
                        metrics = await dashboard_cache.get_metrics(user_id, token)
                        if metrics and metrics.get("focus"):
                            await websocket.send_json({
                                "type": "focus_data",
                                "data": {"focus": metrics["focus"]},
                                "timestamp": datetime.utcnow().isoformat()
                            })
                elif message_type == "get_metrics":
                    # Client is explicitly requesting metrics
                    from data_layer.cache.dashboard_cache import dashboard_cache
                    metrics = await dashboard_cache.get_metrics(user_id, token)
                    await websocket.send_json({
                        "type": "metrics_update",
                        "data": metrics,
                        "timestamp": datetime.utcnow().isoformat()
                    })
                    logger.debug(f"Sent requested metrics to user {user_id}")
                elif message_type == "cache_invalidated_ack":
                    # Client acknowledged cache invalidation and is requesting fresh data
                    logger.info(
                        f"Client acknowledged cache invalidation, fetching fresh metrics for user {user_id}")
                    from data_layer.cache.dashboard_cache import dashboard_cache
                    try:
                        # Force cache refresh and fetch fresh metrics
                        await dashboard_cache.invalidate_cache(user_id)
                        metrics = await dashboard_cache.get_metrics(user_id, token)
                        await websocket.send_json({
                            "type": "fresh_metrics",
                            "data": metrics,
                            "timestamp": datetime.utcnow().isoformat()
                        })
                        logger.info(
                            f"Sent fresh metrics to user {user_id} after cache invalidation")
                    except Exception as e:
                        logger.error(
                            f"Error fetching fresh metrics after cache invalidation: {e}")
                        await websocket.send_json({
                            "type": "error",
                            "message": "Failed to fetch fresh metrics",
                            "timestamp": datetime.utcnow().isoformat()
                        })
                elif message_type == "dashboard_update_ack":
                    # Client acknowledged dashboard update and is requesting fresh data
                    logger.info(
                        f"Client requesting fresh data after dashboard update for user {user_id}")
                    from data_layer.cache.dashboard_cache import dashboard_cache
                    try:
                        # Force cache refresh to ensure we get the latest data
                        await dashboard_cache.invalidate_cache(user_id)
                        # Fetch fresh metrics
                        metrics = await dashboard_cache.get_metrics(user_id, token)
                        await websocket.send_json({
                            "type": "fresh_metrics",
                            "data": metrics,
                            "timestamp": datetime.utcnow().isoformat()
                        })
                        logger.info(
                            f"Sent fresh metrics after dashboard update to user {user_id}")
                    except Exception as e:
                        logger.error(
                            f"Error fetching fresh metrics after dashboard update: {e}")
                        await websocket.send_json({
                            "type": "error",
                            "message": "Failed to fetch fresh metrics",
                            "timestamp": datetime.utcnow().isoformat()
                        })
            except json.JSONDecodeError:
                logger.warning(f"Received invalid JSON from client: {user_id}")
            except Exception as e:
                logger.error(
                    f"Error handling client message: {str(e)}", exc_info=True)

    except WebSocketDisconnect:
        logger.info(f"WebSocket disconnected for user_id: {user_id}")
        await dashboard_ws_manager.disconnect(websocket, user_id)
    except Exception as e:
        logger.error(
            f"WebSocket error for user_id {user_id}: {str(e)}", exc_info=True)
        logger.error(f"WebSocket error traceback: {traceback.format_exc()}")
        await dashboard_ws_manager.disconnect(websocket, user_id)
        try:
            await websocket.close(code=status.WS_1011_INTERNAL_ERROR)
        except:
            pass


async def process_ai_option(
    websocket: WebSocket,
    message_data: Dict[str, Any],
    user_id: str,
    token: str
):
    """Process an AI option selected by the user.
    Uses the agent orchestrator to process the option.
    """
    option_id = message_data.get("option_id")
    target_type = message_data.get("target_type")
    target_id = message_data.get("target_id")
    target_data = message_data.get("target_data")
    try:
        logger.info(
            f"AI process request for option {option_id} on {target_type} {target_id} from user {user_id}")

        # Initialize the orchestrator with Atomic Agents pattern
        logger.info(
            f"Initializing AgentOrchestrator for processing option {option_id}")
        from ai_services.agents.orchestrator import AgentOrchestrator
        orchestrator = AgentOrchestrator()

        # Process the option
        logger.info(
            f"Processing option {option_id} for {target_type} {target_id}")
        result = await orchestrator.process_option(
            option_id=str(option_id),
            target_type=str(target_type),
            target_id=str(target_id),
            user_id=user_id,
            target_data=target_data,
            token=token
        )

        # Check if result is valid
        if not result:
            error_msg = "No result returned from agent"
            logger.error(error_msg)
            await websocket.send_json({
                "type": "ai_option_result",
                "data": {
                    "targetId": target_id,
                    "targetType": target_type,
                    "optionId": option_id,
                    "error": error_msg,
                    "success": False
                }
            })
            return

        logger.info(f"Successfully processed option {option_id}")

        # Send result to client
        await websocket.send_json({
            "type": "ai_option_result",
            "data": {
                "targetId": target_id,
                "targetType": target_type,
                "optionId": option_id,
                "result": result,
                "success": True
            }
        })
    except Exception as e:
        error_msg = f"Error processing AI option: {str(e)}"
        logger.error(error_msg, exc_info=True)

        # Send error to client
        try:
            await websocket.send_json({
                "type": "ai_option_result",
                "data": {
                    "targetId": target_id,
                    "targetType": target_type,
                    "optionId": option_id,
                    "error": error_msg,
                    "success": False
                }
            })
        except Exception as send_err:
            logger.error(f"Error sending error response: {str(send_err)}")


@router.websocket("/ws/dashboard/admin")
async def admin_dashboard_websocket(websocket: WebSocket):
    """
    Admin WebSocket endpoint for monitoring dashboard connections.
    Requires a valid JWT token with admin privileges.
    """
    try:
        token_data = await get_token_from_websocket(websocket)
        if not token_data:
            return

        user = get_current_user(token_data)
        # A proper role check should be implemented.
        if "admin" not in user.get("roles", []):
            logger.warning(
                f"Non-admin user {user.get('user_id')} attempted to connect to admin dashboard.")
            await websocket.close(code=status.WS_1008_POLICY_VIOLATION)
            return

        user_id = user["user_id"]

        # Accept the connection
        await websocket.accept()

        # Send initial stats
        await websocket.send_json({
            "type": "stats",
            "timestamp": datetime.utcnow().isoformat(),
            "data": dashboard_ws_manager.get_stats()
        })

        # Keep the connection alive and periodically send stats
        while True:
            # Wait for messages or timeout
            try:
                data = await websocket.receive_text()
                # If admin requests stats update
                await websocket.send_json({
                    "type": "stats",
                    "timestamp": datetime.utcnow().isoformat(),
                    "data": dashboard_ws_manager.get_stats()
                })
            except WebSocketDisconnect:
                break

    except Exception as e:
        logger.error(f"Admin WebSocket error: {str(e)}")
        try:
            await websocket.close(code=status.WS_1011_INTERNAL_ERROR)
        except:
            pass


async def handle_ai_options_request(websocket: WebSocket, data: Dict[str, Any], user_id: str, token: str):
    """Handle AI options request from client.
    Gets AI options for a target from the agent orchestrator.
    """
    try:
        target_type = data.get("target_type", "")
        target_id = data.get("target_id", "")
        target_data = data.get("target_data", {})

        # Validate required fields
        if not target_type or not target_id:
            logger.error("Missing required fields for AI options request")
            await websocket.send_json({
                "type": "ai_options_response",
                "data": {
                    "targetId": target_id,
                    "targetType": target_type,
                    "error": "Missing required fields (target_type or target_id)",
                    "success": False
                }
            })
            return

        # Ensure target_type is a string
        if not isinstance(target_type, str):
            target_type = str(target_type)

        # Ensure target_id is a string
        if not isinstance(target_id, str):
            target_id = str(target_id)

        # Ensure the target data contains required fields for MCP tools
        enhanced_target_data = dict(target_data) if isinstance(
            target_data, dict) else {}

        # Make sure it has an ID
        if "id" not in enhanced_target_data and target_id:
            enhanced_target_data["id"] = target_id

        # Ensure it has a user_id
        if "user_id" not in enhanced_target_data and user_id:
            enhanced_target_data["user_id"] = user_id

        logger.info(
            f"Target data for {target_type}: {str(enhanced_target_data)[:200]}...")
        logger.info(f"Requesting options from orchestrator for {target_type}")

        # Check if MCP client is available
        mcp_client = get_mcp_client()
        if not mcp_client:
            error_msg = "MCP client not initialized. AI services are currently unavailable."
            logger.error(error_msg)
            await websocket.send_json({
                "type": "ai_options_response",
                "data": {
                    "targetId": target_id,
                    "targetType": target_type,
                    "error": error_msg,
                    "success": False
                }
            })
            return

        # Initialize the orchestrator using Atomic Agents pattern
        try:
            from ai_services.agents.orchestrator import AgentOrchestrator
            orchestrator = AgentOrchestrator()
        except Exception as orchestrator_error:
            error_msg = f"Failed to initialize AI orchestrator: {str(orchestrator_error)}"
            logger.error(error_msg)
            await websocket.send_json({
                "type": "ai_options_response",
                "data": {
                    "targetId": target_id,
                    "targetType": target_type,
                    "error": error_msg,
                    "success": False
                }
            })
            return

        # Get options from orchestrator using the same interface
        try:
            options = await orchestrator.get_options_for_target(
                target_type=target_type,
                target_id=target_id,
                target_data=enhanced_target_data,
                user_id=user_id,
                token=token
            )
        except Exception as options_error:
            error_msg = f"Failed to get AI options: {str(options_error)}"
            logger.error(error_msg)
            await websocket.send_json({
                "type": "ai_options_response",
                "data": {
                    "targetId": target_id,
                    "targetType": target_type,
                    "error": error_msg,
                    "success": False
                }
            })
            return

        logger.info(
            f"Got {len(options)} options from orchestrator for {target_type}")

        # Send options back to client
        await websocket.send_json({
            "type": "ai_options_response",
            "data": {
                "targetId": target_id,
                "targetType": target_type,
                "options": options,
                "success": True
            }
        })
        logger.info(f"Sent ai_options_response for {target_type} {target_id}")

    except Exception as e:
        logger.error(
            f"Error handling AI options request: {str(e)}", exc_info=True)
        # Try to send error response if possible
        try:
            target_type = str(data.get("target_type", "unknown"))
            target_id = str(data.get("target_id", "unknown"))
            await websocket.send_json({
                "type": "ai_options_response",
                "data": {
                    "targetId": target_id,
                    "targetType": target_type,
                    "error": f"Internal server error: {str(e)}",
                    "success": False
                }
            })
        except Exception as send_error:
            logger.error(f"Error sending error response: {str(send_error)}")
