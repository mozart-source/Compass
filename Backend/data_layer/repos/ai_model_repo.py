from typing import List, Optional, Dict, Any, Union
from data_layer.repos.base_repo import BaseMongoRepository
from data_layer.models.ai_model import AIModel, ModelUsage, ModelType, ModelProvider, BillingType
import logging
from datetime import datetime, timedelta

logger = logging.getLogger(__name__)


class AIModelRepository(BaseMongoRepository[AIModel]):
    """Repository for managing AI models in MongoDB."""

    def __init__(self):
        """Initialize the repository with the AIModel model."""
        super().__init__(AIModel)

    def find_by_name_version(self, name: str, version: str) -> Optional[AIModel]:
        """Find model by name and version."""
        return self.find_one({"name": name, "version": version})

    def find_active_models(self) -> List[AIModel]:
        """Find all active models."""
        return self.find_many({"status": "active"})

    def find_by_type(self, model_type: ModelType) -> List[AIModel]:
        """Find models by type."""
        return self.find_many({"type": model_type.value})

    def find_by_provider(self, provider: ModelProvider) -> List[AIModel]:
        """Find models by provider."""
        return self.find_many({"provider": provider.value})

    def create_model(
        self,
        name: str,
        version: str,
        type: ModelType,
        provider: ModelProvider,
        status: str = "active",
        capabilities: Optional[Dict[str, Any]] = None,
        config: Optional[Dict[str, Any]] = None,
        billing_type: Optional[BillingType] = None,
        input_token_cost_per_million: float = 0.0,
        output_token_cost_per_million: float = 0.0,
        provisioned_capacity: Optional[int] = None,
        provisioned_cost_per_hour: Optional[float] = None,
        quota_limit: Optional[int] = None,
        quota_reset_interval: Optional[str] = None
    ) -> AIModel:
        """Create a new AI model."""
        model = AIModel(
            name=name,
            version=version,
            type=type,
            provider=provider,
            status=status,
            capabilities=capabilities or {},
            config=config or {},
            metrics={},
            billing_type=billing_type or BillingType.PAY_AS_YOU_GO,
            input_token_cost_per_million=input_token_cost_per_million,
            output_token_cost_per_million=output_token_cost_per_million,
            provisioned_capacity=provisioned_capacity,
            provisioned_cost_per_hour=provisioned_cost_per_hour,
            quota_limit=quota_limit,
            quota_reset_interval=quota_reset_interval
        )

        model_id = self.insert(model)
        logger.info(f"Created AI model with ID {model_id}")

        return model

    def update_metrics(self, model_id: str, metrics: Dict[str, Any]) -> Optional[AIModel]:
        """Update model metrics."""
        return self.update(model_id, {"metrics": metrics})

    def deactivate_model(self, model_id: str) -> Optional[AIModel]:
        """Deactivate a model."""
        return self.update(model_id, {"status": "inactive"})

    def delete_by_name_version(self, name: str, version: str) -> int:
        """Delete models by name and version."""
        return self.delete_many({"name": name, "version": version})

    # Async methods

    async def async_find_by_name_version(self, name: str, version: str) -> Optional[AIModel]:
        """Find model by name and version (async)."""
        return await self.async_find_one({"name": name, "version": version})

    async def async_find_active_models(self) -> List[AIModel]:
        """Find all active models (async)."""
        return await self.async_find_many({"status": "active"})

    async def async_create_model(
        self,
        name: str,
        version: str,
        type: ModelType,
        provider: ModelProvider,
        status: str = "active",
        capabilities: Optional[Dict[str, Any]] = None,
        config: Optional[Dict[str, Any]] = None,
        billing_type: Optional[BillingType] = None,
        input_token_cost_per_million: float = 0.0,
        output_token_cost_per_million: float = 0.0,
        provisioned_capacity: Optional[int] = None,
        provisioned_cost_per_hour: Optional[float] = None,
        quota_limit: Optional[int] = None,
        quota_reset_interval: Optional[str] = None
    ) -> AIModel:
        """Create a new AI model (async)."""
        model = AIModel(
            name=name,
            version=version,
            type=type,
            provider=provider,
            status=status,
            capabilities=capabilities or {},
            config=config or {},
            metrics={},
            billing_type=billing_type or BillingType.PAY_AS_YOU_GO,
            input_token_cost_per_million=input_token_cost_per_million,
            output_token_cost_per_million=output_token_cost_per_million,
            provisioned_capacity=provisioned_capacity,
            provisioned_cost_per_hour=provisioned_cost_per_hour,
            quota_limit=quota_limit,
            quota_reset_interval=quota_reset_interval
        )

        model_id = await self.async_insert(model)
        logger.info(f"Created AI model with ID {model_id}")

        return model


class ModelUsageRepository(BaseMongoRepository[ModelUsage]):
    """Repository for managing model usage statistics in MongoDB."""

    def __init__(self):
        """Initialize the repository with the ModelUsage model."""
        super().__init__(ModelUsage)

    def log_usage(
        self,
        model_id: str,
        model_name: str,
        request_type: str,
        tokens_in: int = 0,
        tokens_out: int = 0,
        latency_ms: int = 0,
        success: bool = True,
        error: Optional[str] = None,
        user_id: Optional[str] = None,
        session_id: Optional[str] = None,
        input_cost: float = 0.0,
        output_cost: float = 0.0,
        total_cost: float = 0.0,
        billing_type: Union[BillingType, str] = BillingType.PAY_AS_YOU_GO,
        quota_applied: bool = False,
        quota_exceeded: bool = False,
        request_id: str = "",
        endpoint: str = "",
        client_ip: Optional[str] = None,
        user_agent: Optional[str] = None,
        organization_id: Optional[str] = None
    ) -> str:
        """Log model usage."""
        # Convert string billing type to enum if needed
        if isinstance(billing_type, str):
            try:
                billing_type = BillingType(billing_type)
            except ValueError:
                billing_type = BillingType.PAY_AS_YOU_GO

        usage = ModelUsage(
            model_id=model_id,
            model_name=model_name,
            request_type=request_type,
            tokens_in=tokens_in,
            tokens_out=tokens_out,
            latency_ms=latency_ms,
            success=success,
            error=error,
            user_id=user_id,
            session_id=session_id,
            input_cost=input_cost,
            output_cost=output_cost,
            total_cost=total_cost,
            billing_type=billing_type,
            quota_applied=quota_applied,
            quota_exceeded=quota_exceeded,
            request_id=request_id,
            endpoint=endpoint,
            client_ip=client_ip,
            user_agent=user_agent,
            organization_id=organization_id
        )

        usage_id = self.insert(usage)
        return usage_id

    def get_usage_by_model(self, model_id: str, limit: int = 100) -> List[ModelUsage]:
        """Get usage statistics for a specific model."""
        return self.find_many(
            filter={"model_id": model_id},
            limit=limit,
            sort=[("created_at", -1)]
        )

    def get_usage_by_user(self, user_id: str, limit: int = 100) -> List[ModelUsage]:
        """Get usage statistics for a specific user."""
        return self.find_many(
            filter={"user_id": user_id},
            limit=limit,
            sort=[("created_at", -1)]
        )

    async def get_user_quota_usage(
        self,
        user_id: str,
        model_id: str,
        interval: str = "monthly"
    ) -> int:
        """Get total token usage for a user within the quota interval."""
        try:
            # Calculate start time based on interval
            now = datetime.utcnow()
            if interval == "daily":
                start_time = now.replace(
                    hour=0, minute=0, second=0, microsecond=0)
            elif interval == "weekly":
                # Start from Monday of current week
                start_time = now - timedelta(days=now.weekday())
                start_time = start_time.replace(
                    hour=0, minute=0, second=0, microsecond=0)
            else:  # monthly
                # Start from first day of current month
                start_time = now.replace(
                    day=1, hour=0, minute=0, second=0, microsecond=0)

            # Query for usage within the interval
            usages = self.find_many({
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

    def get_user_cost_summary(
        self,
        user_id: str,
        start_time: datetime,
        end_time: datetime
    ) -> Dict[str, Any]:
        """Get cost summary for a user within a time period."""
        try:
            usages = self.find_many({
                "user_id": user_id,
                "created_at": {
                    "$gte": start_time,
                    "$lte": end_time
                }
            })

            summary = {
                "total_cost": sum(usage.total_cost for usage in usages),
                "total_tokens": sum(usage.tokens_in + usage.tokens_out for usage in usages),
                "successful_requests": sum(1 for usage in usages if usage.success),
                "failed_requests": sum(1 for usage in usages if not usage.success),
                "models_used": list(set(usage.model_name for usage in usages))
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

    # Async methods

    async def async_log_usage(
        self,
        model_id: str,
        model_name: str,
        request_type: str,
        tokens_in: int = 0,
        tokens_out: int = 0,
        latency_ms: int = 0,
        success: bool = True,
        error: Optional[str] = None,
        user_id: Optional[str] = None,
        session_id: Optional[str] = None,
        input_cost: float = 0.0,
        output_cost: float = 0.0,
        total_cost: float = 0.0,
        billing_type: Union[BillingType, str] = BillingType.PAY_AS_YOU_GO,
        quota_applied: bool = False,
        quota_exceeded: bool = False,
        request_id: str = "",
        endpoint: str = "",
        client_ip: Optional[str] = None,
        user_agent: Optional[str] = None,
        organization_id: Optional[str] = None
    ) -> str:
        """Log model usage (async)."""
        # Convert string billing type to enum if needed
        if isinstance(billing_type, str):
            try:
                billing_type = BillingType(billing_type)
            except ValueError:
                billing_type = BillingType.PAY_AS_YOU_GO

        usage = ModelUsage(
            model_id=model_id,
            model_name=model_name,
            request_type=request_type,
            tokens_in=tokens_in,
            tokens_out=tokens_out,
            latency_ms=latency_ms,
            success=success,
            error=error,
            user_id=user_id,
            session_id=session_id,
            input_cost=input_cost,
            output_cost=output_cost,
            total_cost=total_cost,
            billing_type=billing_type,
            quota_applied=quota_applied,
            quota_exceeded=quota_exceeded,
            request_id=request_id,
            endpoint=endpoint,
            client_ip=client_ip,
            user_agent=user_agent,
            organization_id=organization_id
        )

        usage_id = await self.async_insert(usage)
        return usage_id
