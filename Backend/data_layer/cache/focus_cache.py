from Backend.data_layer.cache.redis_client import get_cached_value, set_cached_value, delete_cached_value
import json


async def cache_focus_stats(user_id: str, stats: dict, ttl: int = 600):
    await set_cached_value(f"focus_stats:{user_id}", json.dumps(stats), ttl)


async def get_cached_focus_stats(user_id: str):
    cached = await get_cached_value(f"focus_stats:{user_id}")
    return json.loads(cached) if cached else None


async def invalidate_focus_cache(user_id: str) -> None:
    await delete_cached_value(f"focus_stats:{user_id}")
