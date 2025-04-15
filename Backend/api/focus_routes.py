from fastapi import APIRouter, Depends, HTTPException, WebSocket, WebSocketDisconnect
from data_layer.repos.focus_repo import FocusSessionRepository, FocusSettingsRepository
from app.schemas.focus_schemas import FocusSessionCreate, FocusSessionStop, FocusSessionResponse, FocusStatsResponse, FocusSettingsUpdate, FocusSettingsResponse
from utils.jwt import extract_user_id_from_token
from data_layer.models.focus_model import FocusSession
from typing import List, Dict
from datetime import timezone, datetime
import asyncio
from fastapi.encoders import jsonable_encoder
from data_layer.cache.pubsub_manager import pubsub_manager
from data_layer.cache.redis_client import redis_client, redis_pubsub_client
import json
import logging

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/focus", tags=["Focus"])
repo = FocusSessionRepository()
settings_repo = FocusSettingsRepository()

# In-memory store for active dashboard WebSocket connections per user
active_focus_ws_connections: Dict[str, List[WebSocket]] = {}


async def broadcast_focus_update(user_id: str, update_type: str, data: dict):
    """
    Broadcast a focus update to all connected WebSocket clients for a user

    Args:
        user_id: The user ID to send the update to
        update_type: The type of update (focus_stats, focus_session_started, etc.)
        data: The data to send
    """
    if user_id not in active_focus_ws_connections:
        return

    message = {
        "type": update_type,
        "data": data,
        "timestamp": datetime.now().isoformat()
    }

    # Convert to JSON-serializable format
    serializable_message = jsonable_encoder(message)

    # Send to all active connections for this user
    disconnected = []
    for i, websocket in enumerate(active_focus_ws_connections[user_id]):
        try:
            await websocket.send_json(serializable_message)
        except Exception as e:
            logger.error(f"Error sending to WebSocket: {e}")
            disconnected.append(i)

    # Clean up disconnected WebSockets
    for i in sorted(disconnected, reverse=True):
        try:
            await active_focus_ws_connections[user_id][i].close()
        except:
            pass
        del active_focus_ws_connections[user_id][i]

    # Remove user entry if all connections are gone
    if not active_focus_ws_connections[user_id]:
        del active_focus_ws_connections[user_id]


@router.post("/start", response_model=FocusSessionResponse)
async def start_focus_session(data: FocusSessionCreate, user_id: str = Depends(extract_user_id_from_token)):
    active = repo.find_active_session(user_id)
    if active:
        raise HTTPException(
            status_code=400, detail="Focus session already active")
    session_obj = FocusSession(
        user_id=user_id,
        **data.model_dump(),
        end_time=None,
        duration=None,
        status="active",
        interruptions=0,
        metadata={}
    )
    inserted_id = repo.insert(session_obj)
    session = repo.find_by_id(inserted_id)
    if not session:
        raise HTTPException(
            status_code=500, detail="Failed to create focus session")
    response = FocusSessionResponse(**session.model_dump())

    # Get updated stats for immediate UI update
    stats = repo.get_stats(user_id)
    # Add user target to stats
    user_settings = settings_repo.get_user_settings(user_id)
    stats["daily_target_seconds"] = user_settings.daily_target_seconds

    # Add daily breakdown for UI visualization
    from data_layer.cache.dashboard_cache import dashboard_cache
    stats["daily_breakdown"] = dashboard_cache._generate_daily_focus_breakdown(
        user_id)
    serializable_stats = jsonable_encoder(stats)

    # Publish to Redis pub/sub for real-time update
    serializable_response = jsonable_encoder(response.model_dump())
    await pubsub_manager.publish(user_id, "focus_session_started", serializable_response)
    await pubsub_manager.publish(user_id, "focus_stats", serializable_stats)

    # Also broadcast directly to WebSocket connections
    await broadcast_focus_update(user_id, "focus_session_started", serializable_response)
    await broadcast_focus_update(user_id, "focus_stats", serializable_stats)

    # Invalidate dashboard cache
    try:
        from data_layer.cache.dashboard_cache import dashboard_cache
        await dashboard_cache.invalidate_cache(user_id)
    except ImportError:
        pass

    return response


@router.post("/stop", response_model=FocusSessionResponse)
async def stop_focus_session(data: FocusSessionStop, user_id: str = Depends(extract_user_id_from_token)):
    active = repo.find_active_session(user_id)
    if not active or not active.id:
        raise HTTPException(status_code=404, detail="No active session")

    # Ensure both datetimes are timezone-aware and in UTC
    end_time_utc = data.end_time.astimezone(timezone.utc)
    start_time_utc = active.start_time
    if start_time_utc.tzinfo is None:
        start_time_utc = start_time_utc.replace(tzinfo=timezone.utc)

    # Calculate duration, ensuring it's not negative
    duration = int((end_time_utc - start_time_utc).total_seconds())

    # we'll use the current UTC time as the end time
    if duration < 0:
        logger.warning(
            f"Negative duration detected: {duration}s. Start: {start_time_utc}, End: {end_time_utc}")
        logger.warning("Using current UTC time as end time instead")
        end_time_utc = datetime.now(timezone.utc)
        duration = int((end_time_utc - start_time_utc).total_seconds())
        # If still negative, set a minimum positive duration
        if duration < 0:
            duration = 60  # Set to 1 minute minimum

    updated = repo.update(
        active.id,
        {
            "end_time": end_time_utc,
            "duration": duration,
            "status": "completed",
            "notes": data.notes
        }
    )
    if not updated:
        raise HTTPException(
            status_code=500, detail="Failed to update focus session")
    response = FocusSessionResponse(**updated.model_dump())

    # Get updated stats for immediate UI update
    stats = repo.get_stats(user_id)
    # Add user target to stats
    user_settings = settings_repo.get_user_settings(user_id)
    stats["daily_target_seconds"] = user_settings.daily_target_seconds

    # Add daily breakdown for UI visualization
    from data_layer.cache.dashboard_cache import dashboard_cache
    stats["daily_breakdown"] = dashboard_cache._generate_daily_focus_breakdown(
        user_id)
    serializable_stats = jsonable_encoder(stats)

    # Publish to Redis pub/sub for real-time update
    serializable_response = jsonable_encoder(response.model_dump())
    await pubsub_manager.publish(user_id, "focus_session_stopped", serializable_response)
    await pubsub_manager.publish(user_id, "focus_stats", serializable_stats)

    # Also broadcast directly to WebSocket connections
    await broadcast_focus_update(user_id, "focus_session_stopped", serializable_response)
    await broadcast_focus_update(user_id, "focus_stats", serializable_stats)

    # Invalidate dashboard cache
    try:
        from data_layer.cache.dashboard_cache import dashboard_cache
        await dashboard_cache.invalidate_cache(user_id)
    except ImportError:
        pass

    return response


@router.get("/sessions", response_model=List[FocusSessionResponse])
def list_sessions(user_id: str = Depends(extract_user_id_from_token)):
    sessions = repo.find_by_user(user_id)
    return [FocusSessionResponse(**s.model_dump()) for s in sessions]


@router.get("/stats", response_model=FocusStatsResponse)
def get_stats(user_id: str = Depends(extract_user_id_from_token)):
    stats = repo.get_stats(user_id)

    # Get user settings to include target in response
    user_settings = settings_repo.get_user_settings(user_id)
    stats["daily_target_seconds"] = user_settings.daily_target_seconds

    return FocusStatsResponse(**stats)


@router.post("/settings", response_model=FocusSettingsResponse)
async def update_settings(data: FocusSettingsUpdate, user_id: str = Depends(extract_user_id_from_token)):
    """Update user focus settings"""
    # Filter out None values to only update provided fields
    update_data = {k: v for k, v in data.model_dump().items() if v is not None}

    if not update_data:
        raise HTTPException(
            status_code=400, detail="No valid settings provided")

    updated_settings = settings_repo.update_settings(user_id, update_data)

    # Get updated stats to reflect new target
    stats = repo.get_stats(user_id)
    stats["daily_target_seconds"] = updated_settings.daily_target_seconds

    # Broadcast the update to WebSocket clients
    asyncio.create_task(broadcast_focus_update(user_id, "focus_settings_updated",
                                               {"settings": updated_settings.model_dump(), "stats": stats}))

    # Invalidate dashboard cache to ensure fresh data
    try:
        from data_layer.cache.dashboard_cache import dashboard_cache
        asyncio.create_task(dashboard_cache.invalidate_cache(user_id))
    except ImportError:
        pass

    return FocusSettingsResponse(**updated_settings.model_dump())


@router.get("/settings", response_model=FocusSettingsResponse)
def get_settings(user_id: str = Depends(extract_user_id_from_token)):
    """Get user focus settings"""
    settings = settings_repo.get_user_settings(user_id)
    return FocusSettingsResponse(**settings.model_dump())


@router.websocket("/ws")
async def focus_ws(websocket: WebSocket):
    await websocket.accept()
    user_id = None

    try:
        # Get authorization token from header
        token = websocket.headers.get("authorization")
        if not token:
            logger.warning("WebSocket connection attempt without token")
            await websocket.close(code=4001)
            return

        # Extract user ID from token
        user_id = extract_user_id_from_token(token)
        logger.info(f"Focus WebSocket connected for user: {user_id}")

        # Add the WebSocket connection to the active connections
        if user_id not in active_focus_ws_connections:
            active_focus_ws_connections[user_id] = []
        active_focus_ws_connections[user_id].append(websocket)

        # Send initial focus stats
        try:
            stats = repo.get_stats(user_id)
            # Add user target to stats
            user_settings = settings_repo.get_user_settings(user_id)
            stats["daily_target_seconds"] = user_settings.daily_target_seconds

            await websocket.send_json({
                "type": "focus_stats",
                "data": stats
            })
        except Exception as e:
            logger.error(f"Error sending initial focus stats: {e}")

        # Keep the connection alive and handle incoming messages
        while True:
            data = await websocket.receive_text()
            try:
                message = json.loads(data)
                # Handle different message types
                message_type = message.get("type", "")

                if message_type == "ping":
                    await websocket.send_json({"type": "pong"})
                elif message_type == "refresh_stats":
                    stats = repo.get_stats(user_id)
                    # Add user target to stats
                    user_settings = settings_repo.get_user_settings(user_id)
                    stats["daily_target_seconds"] = user_settings.daily_target_seconds

                    await websocket.send_json({
                        "type": "focus_stats",
                        "data": stats
                    })
            except json.JSONDecodeError:
                logger.warning(f"Received invalid JSON from client: {user_id}")
            except Exception as e:
                logger.error(f"Error handling client message: {e}")

    except WebSocketDisconnect:
        logger.info(f"Focus WebSocket disconnected for user: {user_id}")
    except Exception as e:
        logger.error(f"Focus WebSocket error: {e}")
        if websocket.client_state.value == 1:  # OPEN
            await websocket.close(code=4002)
    finally:
        # Clean up connection when done
        if user_id in active_focus_ws_connections:
            if websocket in active_focus_ws_connections[user_id]:
                active_focus_ws_connections[user_id].remove(websocket)
            # Remove the user's entry if there are no more connections
            if not active_focus_ws_connections[user_id]:
                del active_focus_ws_connections[user_id]
