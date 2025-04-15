from starlette.middleware.base import BaseHTTPMiddleware
from starlette.requests import Request
from starlette.responses import Response
import logging
import time
import redis.asyncio as redis
from Backend.core.config import settings

class RateLimiterMiddleware(BaseHTTPMiddleware):
    """
    Middleware to limit API request rate per user/IP using Redis.
    """

    def __init__(self, app, max_requests: int = 100, time_window: int = 60):
        super().__init__(app)
        self.max_requests = max_requests
        self.time_window = time_window  # Time window in seconds

    async def dispatch(self, request: Request, call_next):
        client_ip = request.client.host
        redis_client = request.app.state.redis

        if not redis_client:
            logging.warning("⚠️ Redis not available - Skipping rate limiting!")
            return await call_next(request)

        # Redis Key for Rate Limiting
        key = f"rate_limit:{client_ip}"
        
        # Increment request count
        current_requests = await redis_client.incr(key)
        
        if current_requests == 1:
            await redis_client.expire(key, self.time_window)

        # Check if rate limit exceeded
        if current_requests > self.max_requests:
            logging.warning(f"🚫 Rate limit exceeded for {client_ip}")
            return Response(
                content="Too Many Requests. Slow down!", 
                status_code=429
            )

        return await call_next(request)
