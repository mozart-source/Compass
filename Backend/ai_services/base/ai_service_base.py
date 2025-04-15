from typing import Dict, Optional
import aiohttp
from core.config import settings
from utils.logging_utils import get_logger
from utils.cache_utils import cache_response

logger = get_logger(__name__)


class AIServiceBase:
    def __init__(self, service_name: str):
        self.service_name = service_name
        self.api_key = getattr(settings, f"{service_name.upper()}_API_KEY")
        self.base_url = getattr(
            settings, f"{service_name.upper()}_API_BASE_URL")
        self.session = None
        self.retry_count = 3
        self.timeout = 30

    async def _get_session(self) -> aiohttp.ClientSession:
        if self.session is None or self.session.closed:
            self.session = aiohttp.ClientSession(
                headers={
                    "Authorization": f"Bearer {self.api_key}",
                    "Content-Type": "application/json"
                },
                timeout=aiohttp.ClientTimeout(total=self.timeout)
            )
        return self.session

    async def _make_request(
        self,
        endpoint: str,
        method: str = "POST",
        data: Optional[Dict] = None,
        params: Optional[Dict] = None
    ) -> Dict:
        """Make API request with retry logic and error handling."""
        session = await self._get_session()
        for attempt in range(self.retry_count):
            try:
                async with session.request(
                    method,
                    f"{self.base_url}/{endpoint}",
                    json=data,
                    params=params
                ) as response:
                    response.raise_for_status()
                    return await response.json()
            except Exception as e:
                logger.error(
                    f"{self.service_name} request failed (attempt {attempt + 1}): {str(e)}")
                if attempt == self.retry_count - 1:
                    raise RuntimeError(
                        f"Failed after {self.retry_count} attempts: {str(e)}")
                continue

        # This should never be reached due to the raise in the loop
        raise RuntimeError("Unexpected error in request handling")

    async def close(self):
        """Close the aiohttp session."""
        if self.session and not self.session.closed:
            await self.session.close()

    def __del__(self):
        """Ensure session is closed on deletion."""
        if self.session and not self.session.closed:
            import asyncio
            asyncio.create_task(self.close())
