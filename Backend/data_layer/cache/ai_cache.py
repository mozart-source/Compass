from Backend.data_layer.cache.redis_client import get_cached_value, set_cached_value, delete_cached_value
from typing import Optional, Dict
import json


async def cache_ai_result(key: str, result: Dict, ttl: int = 3600) -> None:
    """Cache AI analysis results."""
    await set_cached_value(f"ai_result:{key}", json.dumps(result), ttl)


async def get_cached_ai_result(key: str) -> Optional[Dict]:
    """Retrieve cached AI analysis results."""
    cached = await get_cached_value(f"ai_result:{key}")
    if not cached:
        return None

    # Ensure we have a valid JSON string, even if empty
    if cached.strip() == "":
        return {}

    try:
        return json.loads(cached)
    except json.JSONDecodeError as e:
        import logging
        logging.error(f"Error decoding cached AI result: {e}")
        return None


async def invalidate_rag_cache(domain: str) -> None:
    """
    Invalidate RAG-related cache entries for a domain (e.g., tasks, meetings).
    """
    pattern = f"rag_cache:*{domain}*"
    await delete_cached_value(pattern)
