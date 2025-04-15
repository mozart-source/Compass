from typing import Dict, Any, Optional, List, Tuple
from datetime import datetime, timedelta
from data_layer.models.ai_model import BillingType
from data_layer.models.base_model import MongoBaseModel
from core.config import settings
import logging
import json
import uuid

logger = logging.getLogger(__name__)


class CostManager:
    """Manages cost tracking and billing for AI services."""

    def __init__(self, mongo_client):
        """Initialize the cost manager."""
        self.mongo_client = mongo_client
        self.enable_quotas = settings.billing_quota_enabled
        self.default_quota_limit = settings.billing_quota_default_limit
        self.quota_reset_interval = settings.billing_quota_reset_interval
        self.cost_tracking_enabled = settings.cost_tracking_enabled
        self.cost_tracking_interval = settings.cost_tracking_interval

    async def calculate_input_cost(self, model_id: str, input_tokens: int) -> float:
        """Calculate the cost for input tokens."""
        try:
            # get_model_by_id is synchronous
            model = self.mongo_client.get_model_by_id(model_id)
            if not model or not hasattr(model, 'input_token_cost_per_million'):
                logger.warning(
                    f"Model {model_id} not found or invalid, using default pricing")
                return 0.0
            input_cost_per_million = model.input_token_cost_per_million
            return (input_tokens / 1_000_000) * input_cost_per_million
        except Exception as e:
            logger.error(f"Error calculating input cost: {str(e)}")
            return 0.0

    async def calculate_output_cost(self, model_id: str, output_tokens: int) -> float:
        """Calculate the cost for output tokens."""
        try:
            # get_model_by_id is synchronous
            model = self.mongo_client.get_model_by_id(model_id)
            if not model or not hasattr(model, 'output_token_cost_per_million'):
                logger.warning(
                    f"Model {model_id} not found or invalid, using default pricing")
                return 0.0
            output_cost_per_million = model.output_token_cost_per_million
            return (output_tokens / 1_000_000) * output_cost_per_million
        except Exception as e:
            logger.error(f"Error calculating output cost: {str(e)}")
            return 0.0

    async def calculate_request_cost(
        self,
        model_id: str,
        input_tokens: int,
        output_tokens: int
    ) -> Tuple[float, float, float]:
        """Calculate the total cost for a request."""
        input_cost = await self.calculate_input_cost(model_id, input_tokens)
        output_cost = await self.calculate_output_cost(model_id, output_tokens)
        total_cost = input_cost + output_cost
        return input_cost, output_cost, total_cost

    async def check_quota(
        self,
        user_id: str,
        model_id: str,
        input_tokens: int,
        output_tokens: int
    ) -> Tuple[bool, str]:
        """Check if the request is within quota limits."""
        if not self.enable_quotas:
            return True, ""

        try:
            # Get user's current usage for this quota period
            current_usage = await self.get_user_quota_usage(
                user_id=user_id,
                model_id=model_id
            )

            # Get model quota limit (fetch synchronously)
            model = self.mongo_client.get_model_by_id(model_id)
            quota_limit = model.quota_limit if model and hasattr(
                model, 'quota_limit') and model.quota_limit is not None else self.default_quota_limit

            # Calculate total tokens for this request
            total_tokens = current_usage + input_tokens + output_tokens

            # Check if quota would be exceeded
            if quota_limit is not None and total_tokens > quota_limit:
                return False, f"Quota exceeded. Limit: {quota_limit}, Current usage: {current_usage}, Requested: {input_tokens + output_tokens}"

            # Check if approaching quota limit
            if quota_limit is not None and total_tokens > quota_limit * settings.billing_quota_alert_threshold:
                await self._send_quota_alert(user_id, total_tokens, quota_limit)

            return True, ""
        except Exception as e:
            logger.error(f"Error checking quota: {str(e)}")
            # On error, allow the request but log the issue
            return True, f"Error checking quota: {str(e)}"

    async def get_user_quota_usage(
        self,
        user_id: str,
        model_id: str
    ) -> int:
        """Get total token usage for a user within the quota interval."""
        try:
            # Calculate start time based on interval
            now = datetime.utcnow()
            if self.quota_reset_interval == "daily":
                start_time = now.replace(
                    hour=0, minute=0, second=0, microsecond=0)
            elif self.quota_reset_interval == "weekly":
                # Start from Monday of current week
                start_time = now - timedelta(days=now.weekday())
                start_time = start_time.replace(
                    hour=0, minute=0, second=0, microsecond=0)
            else:  # monthly
                # Start from first day of current month
                start_time = now.replace(
                    day=1, hour=0, minute=0, second=0, microsecond=0)

            # Query for usage within the interval
            usages = self.mongo_client.model_usage_repo.find_many({
                "user_id": user_id,
                "model_id": model_id,
                "created_at": {"$gte": start_time},
                "success": True
            })

            # Sum up total tokens
            total_tokens = sum(usage.tokens_in +
                               usage.tokens_out for usage in usages)
            return total_tokens

        except Exception as e:
            logger.error(f"Error getting user quota usage: {str(e)}")
            return 0

    async def _send_quota_alert(
        self,
        user_id: str,
        current_usage: int,
        quota_limit: int
    ) -> None:
        """Send alert when approaching quota limit."""
        try:
            # Get user email from user service
            # This would need to be implemented based on your user service
            alert_emails = settings.billing_quota_alert_emails

            # Calculate usage percentage
            usage_percent = (current_usage / quota_limit) * 100

            # Prepare alert message
            message = {
                "type": "quota_alert",
                "user_id": user_id,
                "current_usage": current_usage,
                "quota_limit": quota_limit,
                "usage_percent": usage_percent,
                "timestamp": datetime.utcnow().isoformat()
            }

            # Log alert
            logger.warning(f"Quota alert: {json.dumps(message)}")

            # TODO: Implement actual alert sending mechanism
            # This could be email, webhook, etc.
            pass

        except Exception as e:
            logger.error(f"Error sending quota alert: {str(e)}")

    async def log_cost_tracking(
        self,
        model_id: str,
        user_id: str,
        input_tokens: int,
        output_tokens: int,
        input_cost: float,
        output_cost: float,
        total_cost: float,
        success: bool,
        request_id: str,
        metadata: Optional[Dict[str, Any]] = None
    ) -> None:
        """Log cost tracking information."""
        if not self.cost_tracking_enabled:
            return

        try:
            # Create cost tracking entry
            tracking_data = {
                "model_id": model_id,
                "user_id": user_id,
                "input_tokens": input_tokens,
                "output_tokens": output_tokens,
                "input_cost": input_cost,
                "output_cost": output_cost,
                "total_cost": total_cost,
                "success": success,
                "request_id": request_id,
                "timestamp": datetime.utcnow(),
                "metadata": metadata or {}
            }

            # Store in MongoDB
            await self.mongo_client.cost_tracking_repo.create_tracking_entry(tracking_data)

            # Check if cost alert should be sent
            if total_cost > settings.cost_tracking_alert_threshold:
                await self._send_cost_alert(user_id, total_cost)

        except Exception as e:
            logger.error(f"Error logging cost tracking: {str(e)}")

    async def _send_cost_alert(
        self,
        user_id: str,
        total_cost: float
    ) -> None:
        """Send alert when cost exceeds threshold."""
        try:
            # Prepare alert message
            message = {
                "type": "cost_alert",
                "user_id": user_id,
                "total_cost": total_cost,
                "threshold": settings.cost_tracking_alert_threshold,
                "timestamp": datetime.utcnow().isoformat()
            }

            # Log alert
            logger.warning(f"Cost alert: {json.dumps(message)}")

            # TODO: Implement actual alert sending mechanism
            pass

        except Exception as e:
            logger.error(f"Error sending cost alert: {str(e)}")

    async def get_cost_summary(
        self,
        user_id: str,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None
    ) -> Dict[str, Any]:
        """Get cost summary for a user within a time period."""
        try:
            # Use default time period if not specified
            if not start_time:
                start_time = datetime.utcnow() - timedelta(days=30)
            if not end_time:
                end_time = datetime.utcnow()

            # Get usage data from MongoDB
            return await self.mongo_client.model_usage_repo.get_user_cost_summary(
                user_id=user_id,
                start_time=start_time,
                end_time=end_time
            )

        except Exception as e:
            logger.error(f"Error getting cost summary: {str(e)}")
            return {
                "total_cost": 0.0,
                "total_tokens": 0,
                "successful_requests": 0,
                "failed_requests": 0,
                "models_used": []
            }
