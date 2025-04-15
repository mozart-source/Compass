from fastapi import APIRouter, Depends, HTTPException, Request, WebSocket, WebSocketDisconnect, Query, status, Header
from fastapi.responses import JSONResponse
from typing import Optional, Dict, Any
from utils.jwt import extract_user_id_from_token
from data_layer.cache.dashboard_cache import dashboard_cache
import logging
import json

# Import WebSocket manager
try:
    from api.websocket.dashboard_ws import dashboard_ws_manager
    from api.websocket.routes import router as websocket_router
except ImportError:
    dashboard_ws_manager = None
    websocket_router = None
    logging.getLogger(__name__).warning(
        "WebSocket manager not available, real-time updates will be disabled")

logger = logging.getLogger(__name__)

dashboard_router = APIRouter(prefix="/dashboard", tags=["dashboard"])

# Include WebSocket routes if available
if websocket_router:
    dashboard_router.include_router(websocket_router)


@dashboard_router.get("/metrics")
async def get_dashboard_metrics(
    request: Request,
    user_id: str = Depends(extract_user_id_from_token),
    authorization: Optional[str] = Header(None)
):
    # Extract token from authorization header if present
    token = None
    if authorization and authorization.startswith("Bearer "):
        token = authorization.replace("Bearer ", "")

    logger.debug(f"Fetching dashboard metrics for user {user_id}")
    metrics = await dashboard_cache.get_metrics(user_id, token or "")
    return JSONResponse(content=metrics)


@dashboard_router.get("/metrics/stats")
async def get_dashboard_metrics_stats(is_admin: bool = True):
    """Get dashboard metrics statistics for monitoring (admin only)"""
    if not is_admin:
        raise HTTPException(status_code=403, detail="Admin access required")
        
    try:
        stats = dashboard_cache.get_metrics_statistics()
        return JSONResponse(content={"success": True, "data": stats})
    except Exception as e:
        logger.error(f"Error getting dashboard metrics statistics: {str(e)}")
        return JSONResponse(content={"success": False, "error": "Failed to get dashboard metrics statistics"})


@dashboard_router.post("/refresh")
async def refresh_dashboard_metrics(
    request: Request,
    user_id: str = Depends(extract_user_id_from_token)
):
    """Force refresh dashboard metrics for the authenticated user"""
    try:
        # Invalidate cache for this user
        await dashboard_cache.invalidate_cache(user_id)
        return JSONResponse(content={"success": True, "message": "Dashboard metrics refreshed"})
    except Exception as e:
        logger.error(f"Error refreshing dashboard metrics: {str(e)}")
        raise HTTPException(
            status_code=500, detail=f"Error refreshing dashboard metrics: {str(e)}")
