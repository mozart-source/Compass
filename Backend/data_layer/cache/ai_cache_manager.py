from typing import List, Dict, Any, Optional
import json
from data_layer.cache.redis_client import get_cached_value, set_cached_value, delete_cached_value
import logging

logger = logging.getLogger(__name__)

class AICacheManager:
    """Manages caching for AI-related data using Redis."""
    
    TOOLS_CACHE_KEY = "ai:tools"
    SYSTEM_PROMPT_CACHE_KEY = "ai:system_prompt"
    CACHE_TTL = 3600  # 1 hour default TTL
    
    @classmethod
    async def get_cached_tools(cls) -> Optional[List[Dict[str, Any]]]:
        """Get cached tools if they exist."""
        try:
            cached_data = await get_cached_value(cls.TOOLS_CACHE_KEY)
            if cached_data:
                return json.loads(cached_data)
            return None
        except Exception as e:
            logger.error(f"Error getting cached tools: {str(e)}")
            return None
    
    @classmethod
    async def set_cached_tools(cls, tools: List[Dict[str, Any]], ttl: int = CACHE_TTL) -> bool:
        """Cache tools with TTL."""
        try:
            return await set_cached_value(cls.TOOLS_CACHE_KEY, json.dumps(tools), ttl)
        except Exception as e:
            logger.error(f"Error caching tools: {str(e)}")
            return False
    
    @classmethod
    async def get_cached_system_prompt(cls) -> Optional[str]:
        """Get cached system prompt if it exists."""
        try:
            return await get_cached_value(cls.SYSTEM_PROMPT_CACHE_KEY)
        except Exception as e:
            logger.error(f"Error getting cached system prompt: {str(e)}")
            return None
    
    @classmethod
    async def set_cached_system_prompt(cls, prompt: str, ttl: int = CACHE_TTL) -> bool:
        """Cache system prompt with TTL."""
        try:
            return await set_cached_value(cls.SYSTEM_PROMPT_CACHE_KEY, prompt, ttl)
        except Exception as e:
            logger.error(f"Error caching system prompt: {str(e)}")
            return False
    
    @classmethod
    async def invalidate_cache(cls) -> None:
        """Invalidate all AI-related cache."""
        try:
            await delete_cached_value(cls.TOOLS_CACHE_KEY)
            await delete_cached_value(cls.SYSTEM_PROMPT_CACHE_KEY)
        except Exception as e:
            logger.error(f"Error invalidating cache: {str(e)}") 