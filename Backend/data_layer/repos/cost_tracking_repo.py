from typing import Dict, Any, Optional, List
from data_layer.repos.base_repo import BaseMongoRepository
from data_layer.models.cost_tracking import CostTrackingEntry
from datetime import datetime
import logging

logger = logging.getLogger(__name__)


class CostTrackingRepository(BaseMongoRepository[CostTrackingEntry]):
    """Repository for managing cost tracking entries in MongoDB."""

    def __init__(self):
        """Initialize the repository with the CostTrackingEntry model."""
        super().__init__(CostTrackingEntry)

    async def create_tracking_entry(
        self,
        tracking_data: Dict[str, Any]
    ) -> str:
        """Create a new cost tracking entry."""
        entry = CostTrackingEntry(
            model_id=tracking_data["model_id"],
            user_id=tracking_data["user_id"],
            input_tokens=tracking_data["input_tokens"],
            output_tokens=tracking_data["output_tokens"],
            input_cost=tracking_data["input_cost"],
            output_cost=tracking_data["output_cost"],
            total_cost=tracking_data["total_cost"],
            success=tracking_data["success"],
            request_id=tracking_data["request_id"],
            timestamp=tracking_data["timestamp"],
            metadata=tracking_data.get("metadata", {})
        )

        entry_id = await self.async_insert(entry)
        logger.info(f"Created cost tracking entry with ID {entry_id}")
        return entry_id

    async def get_user_cost_summary(
        self,
        user_id: str,
        start_time: datetime,
        end_time: datetime
    ) -> Dict[str, Any]:
        """Get cost summary for a user within a time period."""
        try:
            entries = await self.async_find_many({
                "user_id": user_id,
                "timestamp": {
                    "$gte": start_time,
                    "$lte": end_time
                }
            })

            summary = {
                "total_cost": sum(entry.total_cost for entry in entries),
                "total_tokens": sum(entry.input_tokens + entry.output_tokens for entry in entries),
                "successful_requests": sum(1 for entry in entries if entry.success),
                "failed_requests": sum(1 for entry in entries if not entry.success),
                "models_used": list(set(entry.model_id for entry in entries))
            }

            return summary
        except Exception as e:
            logger.error(f"Error getting user cost summary: {str(e)}")
            return {
                "total_cost": 0.0,
                "total_tokens": 0,
                "successful_requests": 0,
                "failed_requests": 0,
                "models_used": []
            }

    async def get_model_cost_summary(
        self,
        model_id: str,
        start_time: datetime,
        end_time: datetime
    ) -> Dict[str, Any]:
        """Get cost summary for a model within a time period."""
        try:
            entries = await self.async_find_many({
                "model_id": model_id,
                "timestamp": {
                    "$gte": start_time,
                    "$lte": end_time
                }
            })

            summary = {
                "total_cost": sum(entry.total_cost for entry in entries),
                "total_tokens": sum(entry.input_tokens + entry.output_tokens for entry in entries),
                "successful_requests": sum(1 for entry in entries if entry.success),
                "failed_requests": sum(1 for entry in entries if not entry.success),
                "unique_users": len(set(entry.user_id for entry in entries))
            }

            return summary
        except Exception as e:
            logger.error(f"Error getting model cost summary: {str(e)}")
            return {
                "total_cost": 0.0,
                "total_tokens": 0,
                "successful_requests": 0,
                "failed_requests": 0,
                "unique_users": 0
            }

    async def get_organization_cost_summary(
        self,
        organization_id: str,
        start_time: datetime,
        end_time: datetime
    ) -> Dict[str, Any]:
        """Get cost summary for an organization within a time period."""
        try:
            entries = await self.async_find_many({
                "metadata.organization_id": organization_id,
                "timestamp": {
                    "$gte": start_time,
                    "$lte": end_time
                }
            })

            summary = {
                "total_cost": sum(entry.total_cost for entry in entries),
                "total_tokens": sum(entry.input_tokens + entry.output_tokens for entry in entries),
                "successful_requests": sum(1 for entry in entries if entry.success),
                "failed_requests": sum(1 for entry in entries if not entry.success),
                "unique_users": len(set(entry.user_id for entry in entries)),
                "models_used": list(set(entry.model_id for entry in entries))
            }

            return summary
        except Exception as e:
            logger.error(f"Error getting organization cost summary: {str(e)}")
            return {
                "total_cost": 0.0,
                "total_tokens": 0,
                "successful_requests": 0,
                "failed_requests": 0,
                "unique_users": 0,
                "models_used": []
            }

    async def get_cost_trends(
        self,
        user_id: Optional[str] = None,
        model_id: Optional[str] = None,
        organization_id: Optional[str] = None,
        interval: str = "daily",
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None
    ) -> List[Dict[str, Any]]:
        """Get cost trends over time."""
        try:
            # Build query filter
            query_filter = {}
            if user_id:
                query_filter["user_id"] = user_id
            if model_id:
                query_filter["model_id"] = model_id
            if organization_id:
                query_filter["metadata.organization_id"] = organization_id
            if start_time:
                query_filter["timestamp"] = {"$gte": start_time}
            if end_time:
                if "timestamp" in query_filter:
                    query_filter["timestamp"]["$lte"] = end_time
                else:
                    query_filter["timestamp"] = {"$lte": end_time}

            # Get all entries
            entries = await self.async_find_many(query_filter)

            # Group by interval
            trends = []
            current_interval = {}
            for entry in sorted(entries, key=lambda x: x.timestamp):
                interval_key = self._get_interval_key(
                    entry.timestamp, interval)
                if interval_key not in current_interval:
                    if current_interval:
                        trends.append(current_interval)
                    current_interval = {
                        "interval": interval_key,
                        "total_cost": 0.0,
                        "total_tokens": 0,
                        "successful_requests": 0,
                        "failed_requests": 0
                    }

                current_interval["total_cost"] += entry.total_cost
                current_interval["total_tokens"] += entry.input_tokens + \
                    entry.output_tokens
                if entry.success:
                    current_interval["successful_requests"] += 1
                else:
                    current_interval["failed_requests"] += 1

            # Add last interval
            if current_interval:
                trends.append(current_interval)

            return trends
        except Exception as e:
            logger.error(f"Error getting cost trends: {str(e)}")
            return []

    def _get_interval_key(self, timestamp: datetime, interval: str) -> str:
        """Get key for grouping by interval."""
        if interval == "hourly":
            return timestamp.strftime("%Y-%m-%d %H:00")
        elif interval == "daily":
            return timestamp.strftime("%Y-%m-%d")
        elif interval == "weekly":
            return f"{timestamp.year}-W{timestamp.isocalendar()[1]}"
        else:  # monthly
            return timestamp.strftime("%Y-%m")
