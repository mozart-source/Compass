from Backend.data_layer.cache.redis_client import get_cached_value, set_cached_value, delete_cached_value
from typing import List, Dict, Optional
import json


async def cache_system_metrics(user_id: str, metrics: List[Dict], ttl: int = 600) -> None:
    await set_cached_value(f"system_metrics:{user_id}", json.dumps(metrics), ttl)


async def get_cached_system_metrics(user_id: str) -> Optional[List[Dict]]:
    cached = await get_cached_value(f"system_metrics:{user_id}")
    return json.loads(cached) if cached else None


async def invalidate_system_metrics_cache(user_id: str) -> None:
    await delete_cached_value(f"system_metrics:{user_id}")
