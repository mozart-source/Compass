from pydantic import BaseModel, Field
from typing import Dict, Any, Optional, List, Union
from datetime import datetime


class WebSocketMessage(BaseModel):
    """Base class for all WebSocket messages"""
    type: str
    timestamp: str = Field(
        default_factory=lambda: datetime.utcnow().isoformat())


class ConnectedMessage(WebSocketMessage):
    """Sent when a client successfully connects"""
    type: str = "connected"
    message: str = "Connected to dashboard updates"


class ErrorMessage(WebSocketMessage):
    """Sent when an error occurs"""
    type: str = "error"
    error: str
    details: Optional[Dict[str, Any]] = None


class PingMessage(WebSocketMessage):
    """Ping message to keep connection alive"""
    type: str = "ping"


class PongMessage(WebSocketMessage):
    """Response to ping message"""
    type: str = "pong"


class RefreshRequestMessage(WebSocketMessage):
    """Client request to refresh dashboard data"""
    type: str = "refresh"


class RefreshInitiatedMessage(WebSocketMessage):
    """Confirmation that refresh has been initiated"""
    type: str = "refresh_initiated"


class MetricsUpdateMessage(WebSocketMessage):
    """Sent when metrics are updated"""
    type: str = "metrics_update"
    data: Dict[str, Any]
    source: Optional[str] = None


class CacheInvalidateMessage(WebSocketMessage):
    """Sent when cache is invalidated"""
    type: str = "cache_invalidate"
    data: Dict[str, Any]


class StatsMessage(WebSocketMessage):
    """Connection statistics for admin dashboard"""
    type: str = "stats"
    data: Dict[str, Any]


# Map of message types to their respective classes
MESSAGE_TYPES = {
    "connected": ConnectedMessage,
    "error": ErrorMessage,
    "ping": PingMessage,
    "pong": PongMessage,
    "refresh": RefreshRequestMessage,
    "refresh_initiated": RefreshInitiatedMessage,
    "metrics_update": MetricsUpdateMessage,
    "cache_invalidate": CacheInvalidateMessage,
    "stats": StatsMessage
}


def parse_message(data: Dict[str, Any]) -> WebSocketMessage:
    """Parse a raw message into the appropriate message class"""
    message_type = data.get("type")
    if message_type and message_type in MESSAGE_TYPES:
        return MESSAGE_TYPES[message_type](**data)
    return WebSocketMessage(**data)
