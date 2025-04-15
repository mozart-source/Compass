"""
JWT authentication handler for WebSocket connections.
"""
import logging
from typing import Dict, Any, Optional, Union
import jwt
from fastapi import Request, WebSocket, HTTPException, status

from core.config import settings
from utils.jwt import decode_token

logger = logging.getLogger(__name__)


async def get_token_from_request(request: Request) -> Optional[Dict[str, Any]]:
    """
    Extract and validate JWT token from request.

    Args:
        request: FastAPI request

    Returns:
        Dict containing token payload and raw token, or None if invalid
    """
    auth_header = request.headers.get("Authorization")

    if not auth_header:
        return None

    try:
        scheme, token = auth_header.split()
        if scheme.lower() != "bearer":
            return None

        payload = jwt.decode(
            token,
            settings.jwt_secret_key,
            algorithms=[settings.jwt_algorithm]
        )

        return {"payload": payload, "raw": token}

    except (jwt.PyJWTError, ValueError) as e:
        logger.error(f"Invalid token: {str(e)}")
        return None


async def get_token_from_websocket(websocket: WebSocket) -> Optional[Dict[str, Any]]:
    """
    Extract and validate JWT token from WebSocket connection.

    Args:
        websocket: FastAPI WebSocket connection

    Returns:
        Dict containing token payload and raw token, or None if invalid
    """
    # Try to get token from query parameters
    token = websocket.query_params.get("token")

    # If not in query params, try to get from headers
    if not token:
        auth_header = websocket.headers.get("Authorization")
        if auth_header:
            try:
                scheme, token = auth_header.split()
                if scheme.lower() != "bearer":
                    return None
            except ValueError:
                return None

    if not token:
        # No token found
        return None

    try:
        payload = jwt.decode(
            token,
            settings.jwt_secret_key,
            algorithms=[settings.jwt_algorithm]
        )

        return {"payload": payload, "raw": token}

    except (jwt.PyJWTError, ValueError) as e:
        logger.error(f"Invalid WebSocket token: {str(e)}")
        return None


def get_current_user(token_data: Dict[str, Any]) -> Dict[str, Any]:
    """
    Extract user information from token data.

    Args:
        token_data: Token data from get_token_from_request

    Returns:
        Dict containing user information
    """
    if not token_data or "payload" not in token_data:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid authentication credentials",
            headers={"WWW-Authenticate": "Bearer"},
        )

    payload = token_data["payload"]

    # Extract user ID from payload
    user_id = payload.get("user_id") or payload.get(
        "userId") or payload.get("sub")

    if not user_id:
        raise HTTPException(
            status_code=status.HTTP_401_UNAUTHORIZED,
            detail="Invalid user information in token",
            headers={"WWW-Authenticate": "Bearer"},
        )

    return {
        "user_id": user_id,
        "username": payload.get("username", ""),
        "email": payload.get("email", ""),
        "roles": payload.get("roles", []),
        "permissions": payload.get("permissions", []),
    }
