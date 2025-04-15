from typing import Optional, List, Dict, Any, Union
import redis.asyncio as redis
from core.config import settings
import logging
import json
from difflib import SequenceMatcher
import asyncio

logger = logging.getLogger(__name__)

# Create Redis connection pool with db=1
redis_client = redis.from_url(
    settings.redis_url,
    decode_responses=True,
    db=1  # Use database 1 instead of default 0
)

logger.info("Initialized Redis client on database 1")


async def get_cached_value(key: str) -> Optional[str]:
    """Get a cached value by key."""
    try:
        logger.debug(f"Getting cached value for key: {key}")
        value = await redis_client.get(key)
        if value:
            logger.debug("Cache hit")
            return value
        logger.debug("Cache miss")
        return None
    except Exception as e:
        logger.error(f"Error getting cached value: {str(e)}", exc_info=True)
        return None


async def set_cached_value(key: str, value: str, ttl: int = 3600) -> bool:
    """Set a cached value with TTL."""
    try:
        logger.debug(f"Setting cache value for key: {key} with TTL: {ttl}")
        await redis_client.set(key, value, ex=ttl)
        logger.debug("Cache value set successfully")
        return True
    except Exception as e:
        logger.error(f"Error setting cached value: {str(e)}", exc_info=True)
        return False


async def delete_cached_value(key: str) -> bool:
    """Delete a cached value."""
    try:
        logger.debug(f"Deleting cached value for key: {key}")
        result = await redis_client.delete(key)
        if result:
            logger.debug("Cache value deleted successfully")
        else:
            logger.debug("Key not found in cache")
        return bool(result)
    except Exception as e:
        logger.error(f"Error deleting cached value: {str(e)}", exc_info=True)
        return False


async def get_keys_by_pattern(pattern: str) -> List[str]:
    """Get all Redis keys matching a pattern."""
    try:
        logger.debug(f"Searching for keys matching pattern: {pattern}")
        keys = []
        async for key in redis_client.scan_iter(match=pattern):
            keys.append(key)
        logger.debug(f"Found {len(keys)} matching keys")
        return keys
    except Exception as e:
        logger.error(f"Error getting keys by pattern: {str(e)}", exc_info=True)
        return []


async def get_cache_stats() -> Dict[str, Any]:
    """Get Redis cache statistics."""
    try:
        logger.debug("Getting cache statistics")
        info = await redis_client.info()
        stats = {
            "total_keys": await redis_client.dbsize(),
            "used_memory": info.get("used_memory_human", "unknown"),
            "connected_clients": info.get("connected_clients", 0),
            "uptime_days": info.get("uptime_in_days", 0)
        }
        logger.debug(f"Cache stats retrieved: {stats}")
        return stats
    except Exception as e:
        logger.error(f"Error getting cache stats: {str(e)}", exc_info=True)
        return {
            "error": str(e),
            "total_keys": 0,
            "used_memory": "unknown",
            "connected_clients": 0,
            "uptime_days": 0
        }


class RedisPubSubClient:
    def __init__(self):
        self.redis_url = settings.redis_url
        self._redis = None

    async def get_redis(self):
        if self._redis is None:
            self._redis = await redis.from_url(self.redis_url, decode_responses=True, db=1)
        return self._redis

    async def subscribe(self, channel_name, callback):
        redis_conn = await self.get_redis()
        pubsub = redis_conn.pubsub()

        # Check if this is a pattern subscription (contains *)
        if '*' in channel_name:
            await pubsub.psubscribe(channel_name)
            logger.info(f"Pattern subscribed to {channel_name}")
            async for message in pubsub.listen():
                if message['type'] == 'pmessage':
                    try:
                        event = json.loads(message['data'])
                        await callback(event)
                    except Exception as e:
                        logger.error(f"Error handling pattern event: {e}")
        else:
            await pubsub.subscribe(channel_name)
            logger.info(f"Subscribed to {channel_name}")
            async for message in pubsub.listen():
                if message['type'] == 'message':
                    try:
                        event = json.loads(message['data'])
                        await callback(event)
                    except Exception as e:
                        logger.error(f"Error handling dashboard event: {e}")

    async def close(self):
        if self._redis:
            await self._redis.close()
            self._redis = None


redis_pubsub_client = RedisPubSubClient()
