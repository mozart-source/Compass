import jwt
from fastapi import Header, HTTPException
from core.config import settings
import logging
import uuid

logger = logging.getLogger(__name__)


def extract_user_id_from_token(authorization: str = Header(...)) -> str:
    if not authorization or not authorization.startswith("Bearer "):
        logger.error("Invalid or missing authorization header")
        raise HTTPException(status_code=401, detail="Invalid or missing token")

    token = authorization.split(" ")[1]
    try:
        logger.debug(f"Attempting to decode token: {token[:20]}...")
        claims = jwt.decode(token, settings.jwt_secret_key,
                            algorithms=[settings.jwt_algorithm])

        # Handle both UUID and string formats for user_id (Go backend compatibility)
        user_id = claims.get("user_id")
        if not user_id:
            # Try alternative field names for compatibility
            user_id = claims.get("userId") or claims.get("sub")

        if not user_id:
            logger.error("Token missing user_id claim")
            raise HTTPException(
                status_code=401, detail="user_id not found in token")

        # Convert UUID to string if necessary
        if isinstance(user_id, dict) and "UUID" in str(user_id):
            # Handle Go backend UUID format
            user_id = str(user_id)
        elif not isinstance(user_id, str):
            user_id = str(user_id)

        logger.debug(f"Successfully decoded token for user_id: {user_id}")
        return user_id
    except jwt.ExpiredSignatureError:
        logger.error("Token has expired")
        raise HTTPException(status_code=401, detail="Token has expired")
    except jwt.InvalidTokenError as e:
        logger.error(f"Invalid token: {str(e)}")
        raise HTTPException(status_code=401, detail=f"Invalid token: {str(e)}")
    except Exception as e:
        logger.error(f"Token decode error: {str(e)}", exc_info=True)
        raise HTTPException(
            status_code=401, detail=f"Token decode error: {str(e)}")


def decode_token(token: str):
    """Decode JWT token without raising exceptions"""
    try:
        logger.debug(f"Attempting to decode token: {token[:20]}...")
        claims = jwt.decode(token, settings.jwt_secret_key,
                            algorithms=[settings.jwt_algorithm])

        # Normalize user_id format for cross-service compatibility
        if "user_id" in claims:
            user_id = claims["user_id"]
            # Convert UUID to string if necessary
            if isinstance(user_id, dict) and "UUID" in str(user_id):
                claims["user_id"] = str(user_id)
            elif not isinstance(user_id, str):
                claims["user_id"] = str(user_id)

        logger.debug(f"Successfully decoded token")
        return claims
    except jwt.ExpiredSignatureError:
        logger.error("Token has expired")
        return None
    except jwt.InvalidTokenError as e:
        logger.error(f"Invalid token: {str(e)}")
        return None
    except Exception as e:
        logger.error(f"Token decode error: {str(e)}", exc_info=True)
        return None
