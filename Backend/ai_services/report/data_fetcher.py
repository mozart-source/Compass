"""
Data fetcher service for gathering data from various sources for report generation.
"""

import logging
import json
import hashlib
from typing import Dict, Any, Optional, List, Union, Mapping, cast
from datetime import datetime, timedelta

from mcp_py.client import MCPClient
from data_layer.cache.redis_client import redis_client
from core.mcp_state import get_mcp_client

logger = logging.getLogger(__name__)


class DataFetcherService:
    """
    Service for fetching data from various sources with caching.

    This service coordinates data gathering from:
    - Go backend services via MCP
    - Redis cache for dashboard data
    - Python backend services directly

    It implements caching to avoid redundant data fetching and
    provides a unified interface for report agents.
    """

    def __init__(self, mcp_client: Optional[MCPClient] = None):
        """
        Initialize the data fetcher service.

        Args:
            mcp_client: Optional MCP client instance. If not provided, the global one will be used.
        """
        self.mcp_client = mcp_client
        self.cache_ttl = 300  # Default cache TTL: 5 minutes

    async def fetch_user_data(
        self,
        user_id: str,
        data_type: str,
        parameters: Dict[str, Any],
        time_range: Optional[Dict[str, str]] = None,
        auth_token: Optional[str] = None,
        force_refresh: bool = False
    ) -> Dict[str, Any]:
        """
        Fetch user data from appropriate source with caching.

        Args:
            user_id: User ID
            data_type: Type of data to fetch (e.g., 'activity', 'tasks', 'habits')
            parameters: Additional parameters for the data fetch
            time_range: Optional time range for data (start_date, end_date)
            auth_token: Optional auth token for authenticated requests
            force_refresh: Whether to bypass cache and force a refresh

        Returns:
            Dict containing the requested data
        """
        # Lazy-load the MCP client if it wasn't provided during initialization
        if self.mcp_client is None:
            self.mcp_client = get_mcp_client()

        # Generate cache key based on parameters
        cache_key = self._generate_cache_key(
            user_id, data_type, parameters, time_range)
        logger.info(f"Using cache key for report data: {cache_key}")

        # Try to get data from cache first (unless force refresh is requested)
        if not force_refresh:
            cached_data = await self._get_cached_data(cache_key)
            if cached_data:
                logger.info(
                    f"CACHE HIT for {data_type} data for user {user_id}")
                return cached_data

        logger.info(
            f"CACHE MISS for {data_type} data for user {user_id}. Fetching from source.")

        # Fetch data based on type
        result = await self._fetch_data_by_type(
            user_id,
            data_type,
            parameters,
            time_range,
            auth_token
        )

        # Cache the result
        if result is not None:
            await self._cache_data(cache_key, result)

        return result

    async def fetch_metrics(
        self,
        user_id: str,
        metric_types: List[str],
        time_range: Dict[str, str],
        auth_token: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Fetch multiple types of metrics for a user.

        Args:
            user_id: User ID
            metric_types: List of metric types to fetch
            time_range: Time range for metrics (start_date, end_date)
            auth_token: Optional auth token for authenticated requests

        Returns:
            Dict containing all requested metrics
        """
        results: Dict[str, Any] = {}

        # Ensure time_range has string values
        safe_time_range: Dict[str, str] = {
            k: v if v is not None else ""
            for k, v in time_range.items()
        }

        for metric_type in metric_types:
            data = await self.fetch_user_data(
                user_id,
                metric_type,
                {},
                safe_time_range,
                auth_token
            )
            results[metric_type] = data

        return results

    async def fetch_dashboard_data(
        self,
        user_id: str,
        auth_token: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Fetch dashboard data for a user.

        Args:
            user_id: User ID
            auth_token: Optional auth token for authenticated requests

        Returns:
            Dict containing dashboard data
        """
        try:
            # Import dashboard_cache
            from data_layer.cache.dashboard_cache import dashboard_cache

            # Call the get_metrics method with user_id and auth_token
            metrics = await dashboard_cache.get_metrics(user_id, auth_token or "")

            # Ensure we return a dictionary
            if metrics is None:
                return {}

            return metrics
        except Exception as e:
            logger.error(f"Error fetching dashboard data: {str(e)}")
            return {}

    async def _fetch_data_by_type(
        self,
        user_id: str,
        data_type: str,
        parameters: Dict[str, Any],
        time_range: Optional[Dict[str, str]],
        auth_token: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Fetch data from the appropriate source based on data type.

        Args:
            user_id: User ID
            data_type: Type of data to fetch
            parameters: Additional parameters
            time_range: Optional time range
            auth_token: Optional auth token

        Returns:
            Dict containing the requested data
        """
        # Set default time range if not provided
        if not time_range:
            end_date = datetime.utcnow()
            # Default to last 30 days
            start_date = end_date - timedelta(days=30)
            time_range = {
                "start_date": start_date.strftime("%Y-%m-%d"),
                "end_date": end_date.strftime("%Y-%m-%d")
            }

        # Ensure time_range values are strings
        safe_time_range: Dict[str, str] = {
            k: v if v is not None else ""
            for k, v in time_range.items()
        }

        # Map data types to MCP methods
        mcp_methods = {
            "activity": "user.getInfo",
            "tasks": "get_tasks",
            "calendar": "calendar.getEvents",
            "habits": "get_items",
            "todos": "get_items",
            "focus": "user.getInfo",
            "productivity": "user.getInfo",
            "workflow": "user.getInfo",
            "projects": "get_projects"
        }

        # If data type is in our map, use MCP to fetch from Go backend
        if data_type in mcp_methods:
            try:
                mcp_params = {
                    "user_id": user_id,
                    **parameters
                }

                # Pass auth token if available for MCP calls
                if auth_token:
                    mcp_params["authorization"] = f"Bearer {auth_token}"

                # Add item_type for get_items tool
                if mcp_methods[data_type] == "get_items":
                    mcp_params["item_type"] = data_type

                # Add time range if applicable, but not for user.getInfo
                if safe_time_range and mcp_methods[data_type] != "user.getInfo":
                    mcp_params.update({
                        "start_date": safe_time_range.get("start_date", ""),
                        "end_date": safe_time_range.get("end_date", "")
                    })

                if not self.mcp_client:
                    raise ConnectionError("MCP client is not available.")

                response: Any = await self.mcp_client.call_tool(
                    mcp_methods[data_type],
                    mcp_params
                )

                if not isinstance(response, dict) or response.get('status') != 'success':
                    error_message = response.get('message', 'Unknown error') if isinstance(
                        response, dict) else str(response)
                    logger.error(
                        f"MCP call for {data_type} failed: {error_message}")
                    return {}

                # The MCP client can return different object types. We need to handle them safely.
                call_tool_result = response.get('content')

                # Check if the tool call itself resulted in an error
                if hasattr(call_tool_result, 'isError') and getattr(call_tool_result, 'isError'):
                    logger.warning(
                        f"MCP tool '{mcp_methods[data_type]}' returned an error response: {getattr(call_tool_result, 'content', 'No content')}")
                    return {}

                # Get the content from the result object
                result_content = getattr(call_tool_result, 'content', None)

                # The actual data is a JSON string inside a TextContent object within a list
                if isinstance(result_content, list) and len(result_content) > 0:
                    text_content_item = result_content[0]
                    if hasattr(text_content_item, 'text') and isinstance(text_content_item.text, str):
                        try:
                            # The text attribute contains the final JSON string
                            return json.loads(text_content_item.text)
                        except (json.JSONDecodeError, TypeError) as e:
                            logger.error(
                                f"Could not parse JSON from TextContent for {data_type}. Error: {e}. Content: {text_content_item.text}")
                            return {}

                # Fallback for unexpected structures
                logger.warning(
                    f"Unexpected content structure for {data_type} from MCP: {type(result_content)}")
                return {}

            except Exception as e:
                logger.error(
                    f"Error fetching {data_type} data via MCP: {str(e)}")
                return {}

        # For dashboard data, use the dashboard cache
        elif data_type == "dashboard":
            return await self.fetch_dashboard_data(user_id, auth_token)

        # For unknown data types, return empty dict
        else:
            logger.warning(f"Unknown data type: {data_type}")
            return {}

    async def _get_cached_data(self, cache_key: str) -> Optional[Dict[str, Any]]:
        """
        Get data from cache if available.

        Args:
            cache_key: Cache key

        Returns:
            Cached data if available, None otherwise
        """
        try:
            data = await redis_client.get(f"report_data:{cache_key}")
            if data:
                return json.loads(data)
        except Exception as e:
            logger.error(f"Error getting cached data: {str(e)}")

        return None

    async def _cache_data(self, cache_key: str, data: Dict[str, Any]) -> None:
        """
        Cache data with the specified key.

        Args:
            cache_key: Cache key
            data: Data to cache
        """
        try:
            # Set cache with expiration atomically
            # Check if data is serializable before caching
            json.dumps(data)
            await redis_client.set(
                f"report_data:{cache_key}",
                json.dumps(data),
                ex=self.cache_ttl
            )
        except TypeError as e:
            logger.error(
                f"Error serializing data for caching (key: {cache_key}): {str(e)}. Data: {data}")
        except Exception as e:
            logger.error(f"Error caching data: {str(e)}")

    def _generate_cache_key(
        self,
        user_id: str,
        data_type: str,
        parameters: Dict[str, Any],
        time_range: Optional[Dict[str, str]]
    ) -> str:
        """
        Generate a cache key based on request parameters.

        Args:
            user_id: User ID
            data_type: Type of data
            parameters: Additional parameters
            time_range: Optional time range

        Returns:
            Cache key string
        """
        # Create a dict with all parameters that affect the data
        key_dict = {
            "user_id": user_id,
            "data_type": data_type,
            "parameters": parameters,
            "time_range": time_range or {}
        }

        # Convert to JSON string and hash
        key_str = json.dumps(key_dict, sort_keys=True)
        return hashlib.md5(key_str.encode()).hexdigest()
