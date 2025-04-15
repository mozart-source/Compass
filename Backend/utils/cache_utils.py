from functools import wraps
from typing import Any, Callable, Optional, Dict, List, Union
import json
from datetime import datetime, timedelta
import redis
import time
import enum
from Backend.core.config import settings
from Backend.utils.logging_utils import get_logger

logger = get_logger(__name__)

# Initialize Redis client
redis_client = redis.Redis(
    host=settings.REDIS_HOST,
    port=settings.REDIS_PORT,
    db=settings.REDIS_DB,
    password=settings.REDIS_PASSWORD,
    decode_responses=True
)

# Cache metrics tracking
CACHE_METRICS = {
    'hits': 0,
    'misses': 0,
    'hit_rate': 0.0,
    'last_reset': time.time(),
    'by_type': {}
}

# Cache TTL constants for different types of data
CACHE_TTL_SETTINGS = {
    # Task-related cache TTLs
    'task': 3600,  # 1 hour for individual tasks
    'task_list': 300,  # 5 minutes for task lists
    'task_details': 1800,  # 30 minutes for detailed task info
    'task_metrics': 600,  # 10 minutes for task metrics
    'task_dependencies': 1200,  # 20 minutes for task dependencies
    'task_history': 3600,  # 1 hour for task history

    # Default TTL
    'default': 3600  # 1 hour default
}


def generate_cache_key(func: Callable, *args, **kwargs) -> str:
    """Generate a unique cache key based on function name and arguments."""
    try:
        # Create a unique key based on function name and arguments
        key_parts = [
            func.__module__,
            func.__name__,
            str(args),
            str(sorted(kwargs.items()))
        ]
        return "::".join(key_parts)
    except Exception as e:
        logger.error(f"Error generating cache key: {str(e)}")
        return f"{func.__module__}::{func.__name__}::fallback"


def generate_entity_cache_key(entity_type: str, entity_id: Union[int, str], action: Optional[str] = None) -> str:
    """Generate a cache key for a specific entity.

    Args:
        entity_type: Type of entity (e.g., 'task', 'user')
        entity_id: ID of the entity
        action: Optional action or view type
    """
    if action:
        return f"{entity_type}::{entity_id}::{action}"
    return f"{entity_type}::{entity_id}"


def serialize_data(data: Any) -> str:
    """Serialize data to JSON string format."""
    try:
        def clean_sqlalchemy_obj(obj):
            if hasattr(obj, '_sa_instance_state'):
                # Get all attributes excluding SQLAlchemy internal ones
                cleaned = {k: v for k, v in obj.__dict__.items()
                           if not k.startswith('_sa_')}
                # Handle dependencies list
                if '_dependencies_list' in cleaned:
                    try:
                        cleaned['dependencies'] = json.loads(cleaned['_dependencies_list']) \
                            if cleaned['_dependencies_list'] else []
                    except (json.JSONDecodeError, TypeError):
                        cleaned['dependencies'] = []
                # Convert datetime objects to ISO format strings and handle enums
                for k, v in cleaned.items():
                    cleaned[k] = process_value(v)
                return cleaned
            return process_value(obj)

        def process_value(value):
            """Process any value to ensure it's JSON serializable."""
            if isinstance(value, datetime):
                return value.isoformat()
            elif isinstance(value, enum.Enum):
                return value.value
            elif isinstance(value, dict):
                return {k: process_value(v) for k, v in value.items()}
            elif isinstance(value, list):
                return [process_value(item) for item in value]
            elif hasattr(value, '_sa_instance_state'):
                return clean_sqlalchemy_obj(value)
            return value

        if hasattr(data, '_sa_instance_state'):
            return json.dumps(clean_sqlalchemy_obj(data))
        elif isinstance(data, list):
            return json.dumps([clean_sqlalchemy_obj(item) if hasattr(item, '_sa_instance_state')
                               else process_value(item) for item in data])
        elif isinstance(data, dict):
            return json.dumps({k: process_value(v) for k, v in data.items()})
        return json.dumps(process_value(data))
    except Exception as e:
        logger.error(f"Error serializing data: {str(e)}")
        raise


def deserialize_data(data_str: str) -> Any:
    """Deserialize JSON string back to Python object."""
    try:
        return json.loads(data_str)
    except Exception as e:
        logger.error(f"Error deserializing data: {str(e)}")
        raise


def cache_response(ttl: Optional[int] = 1800, cache_type: str = 'default'):
    """Decorator to cache function responses in Redis.

    Args:
        ttl (int, optional): Time to live in seconds for cached data. If None, uses the cache_type setting.
        cache_type (str): Type of cache to determine TTL if ttl is not provided.
    """
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        async def wrapper(*args, **kwargs):
            try:
                # Determine the TTL to use
                cache_ttl = ttl if ttl is not None else CACHE_TTL_SETTINGS.get(
                    cache_type, CACHE_TTL_SETTINGS['default'])

                # Generate cache key
                cache_key = generate_cache_key(func, *args, **kwargs)

                # Try to get cached response
                cached_data = redis_client.get(cache_key)
                if cached_data:
                    # Track cache hit
                    track_cache_event(hit=True, cache_type=cache_type)
                    logger.info(
                        f"Cache hit for key: {cache_key} [type: {cache_type}]")
                    return deserialize_data(cached_data)

                # If no cache, execute function and cache result
                # Track cache miss
                track_cache_event(hit=False, cache_type=cache_type)
                logger.info(
                    f"Cache miss for key: {cache_key} [type: {cache_type}]")
                result = await func(*args, **kwargs)

                # Don't cache None results
                if result is None:
                    return None

                serialized_result = serialize_data(result)

                # Store in Redis with TTL
                redis_client.setex(cache_key, cache_ttl, serialized_result)

                return result
            except redis.RedisError as e:
                logger.error(f"Redis error in cache_response: {str(e)}")
                # Fall back to executing function without caching
                return await func(*args, **kwargs)
            except Exception as e:
                logger.error(f"Unexpected error in cache_response: {str(e)}")
                raise

        return wrapper
    return decorator


def invalidate_cache(entity_type: str, entity_id: Union[int, str]) -> None:
    """Invalidate all cache entries for a specific entity.

    Args:
        entity_type (str): Type of entity (e.g., 'task', 'user')
        entity_id (int or str): ID of the entity
    """
    try:
        pattern = f"{entity_type}::{entity_id}*"
        clear_cache(pattern)
        logger.debug(f"Invalidated cache for {entity_type} {entity_id}")
    except Exception as e:
        logger.error(f"Error invalidating cache: {str(e)}")


def clear_cache(pattern: str = "*") -> None:
    """Clear cache entries matching the given pattern.

    Args:
        pattern (str): Pattern to match cache keys. Defaults to all keys.
    """
    try:
        cursor = 0
        while True:
            cursor, keys = redis_client.scan(cursor, match=pattern)
            if keys:
                redis_client.delete(*keys)
                logger.debug(
                    f"Cleared {len(keys)} cache keys matching pattern: {pattern}")
            if cursor == 0:
                break
    except Exception as e:
        logger.error(f"Error clearing cache: {str(e)}")
        raise


def get_cache_stats() -> dict:
    """Get cache statistics and metrics."""
    try:
        info = redis_client.info()
        redis_hit_rate = info.get(
            "keyspace_hits", 0) / (info.get("keyspace_hits", 0) + info.get("keyspace_misses", 1) or 1)

        # Calculate application-level hit rate
        app_hit_rate = CACHE_METRICS['hits'] / \
            (CACHE_METRICS['hits'] + CACHE_METRICS['misses'] or 1)

        # Get per-type metrics
        type_metrics = {}
        for cache_type, metrics in CACHE_METRICS['by_type'].items():
            type_hit_rate = metrics['hits'] / \
                (metrics['hits'] + metrics['misses'] or 1)
            type_metrics[cache_type] = {
                'hits': metrics['hits'],
                'misses': metrics['misses'],
                'hit_rate': round(type_hit_rate * 100, 2)
            }

        return {
            "redis": {
                "used_memory": info.get("used_memory_human"),
                "connected_clients": info.get("connected_clients"),
                "total_keys": redis_client.dbsize(),
                "uptime_days": info.get("uptime_in_days"),
                "hit_rate": round(redis_hit_rate * 100, 2)
            },
            "application": {
                "hits": CACHE_METRICS['hits'],
                "misses": CACHE_METRICS['misses'],
                "hit_rate": round(app_hit_rate * 100, 2),
                "tracking_since": datetime.fromtimestamp(CACHE_METRICS['last_reset']).isoformat(),
                "by_type": type_metrics
            }
        }
    except Exception as e:
        logger.error(f"Error getting cache stats: {str(e)}")
        raise


def reset_cache_metrics():
    """Reset the cache hit/miss metrics."""
    global CACHE_METRICS
    CACHE_METRICS = {
        'hits': 0,
        'misses': 0,
        'hit_rate': 0.0,
        'last_reset': time.time(),
        'by_type': {}
    }
    logger.info("Cache metrics have been reset")


def track_cache_event(hit: bool, cache_type: str = 'default'):
    """Track a cache hit or miss event."""
    global CACHE_METRICS

    # Update global counters
    if hit:
        CACHE_METRICS['hits'] += 1
    else:
        CACHE_METRICS['misses'] += 1

    # Update hit rate
    total = CACHE_METRICS['hits'] + CACHE_METRICS['misses']
    CACHE_METRICS['hit_rate'] = CACHE_METRICS['hits'] / \
        total if total > 0 else 0

    # Initialize cache type if not exists
    if cache_type not in CACHE_METRICS['by_type']:
        CACHE_METRICS['by_type'][cache_type] = {'hits': 0, 'misses': 0}

    # Update type-specific counters
    if hit:
        CACHE_METRICS['by_type'][cache_type]['hits'] += 1
    else:
        CACHE_METRICS['by_type'][cache_type]['misses'] += 1


def cache_entity(entity_type: str, ttl: Optional[int] = None):
    """Decorator to cache entity responses with automatic key generation.

    Args:
        entity_type (str): Type of entity for determining TTL and key prefix
        ttl (int, optional): Override the default TTL for this entity type
    """
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        async def wrapper(*args, **kwargs):
            try:
                # Extract entity_id from the first argument after self
                if len(args) >= 2:
                    entity_id = args[1]  # args[0] is typically 'self'
                elif 'id' in kwargs:
                    entity_id = kwargs['id']
                elif f"{entity_type}_id" in kwargs:
                    entity_id = kwargs[f"{entity_type}_id"]
                else:
                    # If we can't determine the entity ID, fall back to standard caching
                    return await cache_response(ttl, entity_type)(func)(*args, **kwargs)

                # Determine the TTL to use
                cache_ttl = ttl if ttl is not None else CACHE_TTL_SETTINGS.get(
                    entity_type, CACHE_TTL_SETTINGS['default'])

                # Generate cache key based on entity type and ID
                cache_key = generate_entity_cache_key(entity_type, entity_id)

                # Try to get cached response
                cached_data = redis_client.get(cache_key)
                if cached_data:
                    # Track cache hit
                    track_cache_event(hit=True, cache_type=entity_type)
                    logger.info(
                        f"Entity cache hit for {entity_type} {entity_id}")
                    return deserialize_data(cached_data)

                # If no cache, execute function and cache result
                # Track cache miss
                track_cache_event(hit=False, cache_type=entity_type)
                logger.info(f"Entity cache miss for {entity_type} {entity_id}")
                result = await func(*args, **kwargs)

                # Don't cache None results
                if result is None:
                    return None

                serialized_result = serialize_data(result)

                # Store in Redis with TTL
                redis_client.setex(cache_key, cache_ttl, serialized_result)

                return result
            except redis.RedisError as e:
                logger.error(f"Redis error in cache_entity: {str(e)}")
                # Fall back to executing function without caching
                return await func(*args, **kwargs)
            except Exception as e:
                logger.error(f"Unexpected error in cache_entity: {str(e)}")
                raise

        return wrapper
    return decorator
