from typing import Dict, Any, List, Optional
from data_layer.mongodb.connection import get_mongodb_client as get_pooled_client, get_async_mongodb_client
from data_layer.models.ai_model import AIModel, ModelType, ModelProvider
from data_layer.models.conversation import Conversation
from data_layer.repos.ai_model_repo import AIModelRepository, ModelUsageRepository
from data_layer.repos.conversation_repo import ConversationRepository
from data_layer.repos.cost_tracking_repo import CostTrackingRepository
import logging
from datetime import datetime
import asyncio

logger = logging.getLogger(__name__)


class MongoDBClient:
    """MongoDB client for AI services."""

    def __init__(self):
        """Initialize MongoDB client."""
        logger.info("Initializing MongoDB client for AI services")

        # Get the pooled MongoDB clients (sync and async)
        self._mongodb_client = get_pooled_client()
        self._async_mongodb_client = get_async_mongodb_client()

        # Initialize repositories
        self._ai_model_repo = AIModelRepository()
        self._model_usage_repo = ModelUsageRepository()
        self._conversation_repo = ConversationRepository()
        self._cost_tracking_repo = CostTrackingRepository()

        # Create a task for initializing collections
        self._init_task = asyncio.create_task(self._ensure_collections_exist())

        # Store the task to prevent it from being garbage collected
        self._tasks = [self._init_task]

    async def _ensure_collections_exist(self):
        """Ensure all collections exist by creating and removing sample documents."""
        try:
            logger.info("Ensuring collections exist in MongoDB")

            # Create sample AIModel through the repository methods
            try:
                # Use existing repository method to create a model
                model = self._ai_model_repo.create_model(
                    name="sample_model",
                    version="0.0.1",
                    provider=ModelProvider.OPENAI,
                    type=ModelType.TEXT_GENERATION,
                    status="inactive"
                )
                if model and model.id:
                    # Delete the sample model
                    self._ai_model_repo.delete(model.id)
                    logger.info("AI models collection initialized")
            except Exception as e:
                logger.error(
                    f"Error initializing AI models collection: {str(e)}")

            # Create sample ModelUsage
            try:
                # Use existing repository method to log usage
                usage_id = self._model_usage_repo.log_usage(
                    model_id="sample",
                    model_name="sample_model",
                    request_type="test",
                    tokens_in=0,
                    tokens_out=0,
                    latency_ms=0,
                    success=True,
                    error=None,
                    user_id="sample_user",
                    session_id="sample_session"
                )
                if usage_id:
                    # Delete the sample usage
                    self._model_usage_repo.delete(usage_id)
                    logger.info("Model usage collection initialized")
            except Exception as e:
                logger.error(
                    f"Error initializing model usage collection: {str(e)}")

            # Create sample Conversation
            try:
                # Use existing repository method to create a conversation
                now = datetime.utcnow()
                conversation = self._conversation_repo.create_conversation(
                    user_id="sample_user",
                    session_id="sample_session",
                    title="Sample Conversation",
                    domain="test"
                )

                if conversation and conversation.id:
                    # Delete the sample conversation
                    self._conversation_repo.delete(conversation.id)
                    logger.info("Conversations collection initialized")
            except Exception as e:
                logger.error(
                    f"Error initializing conversations collection: {str(e)}")

            # Create sample CostTrackingEntry
            try:
                # Use existing repository method to create a tracking entry
                tracking_data = {
                    "model_id": "sample_model",
                    "user_id": "sample_user",
                    "input_tokens": 0,
                    "output_tokens": 0,
                    "input_cost": 0.0,
                    "output_cost": 0.0,
                    "total_cost": 0.0,
                    "success": True,
                    "request_id": "sample_request",
                    "timestamp": datetime.utcnow(),
                    "metadata": {}
                }
                # Create and immediately delete the sample entry
                await self._cost_tracking_repo.create_tracking_entry(tracking_data)
                logger.info("Cost tracking collection initialized")
            except Exception as e:
                logger.error(
                    f"Error initializing cost tracking collection: {str(e)}")

            logger.info("All collections successfully initialized")
        except Exception as e:
            logger.error(f"Error ensuring collections exist: {str(e)}")
            # Don't raise exception to allow startup to continue

    @property
    def ai_model_repo(self) -> AIModelRepository:
        """Get AI model repository."""
        return self._ai_model_repo

    @property
    def model_usage_repo(self) -> ModelUsageRepository:
        """Get model usage repository."""
        return self._model_usage_repo

    @property
    def conversation_repo(self) -> ConversationRepository:
        """Get conversation repository."""
        return self._conversation_repo

    @property
    def cost_tracking_repo(self) -> CostTrackingRepository:
        """Get cost tracking repository."""
        return self._cost_tracking_repo

    # Convenience methods for AI models

    def get_model_by_name_version(self, name: str, version: str) -> Optional[AIModel]:
        """Get AI model by name and version."""
        return self.ai_model_repo.find_by_name_version(name, version)

    def get_active_models(self) -> List[AIModel]:
        """Get all active AI models."""
        return self.ai_model_repo.find_active_models()

    def get_model_by_id(self, model_id: str) -> Optional[AIModel]:
        """Get AI model by its ID."""
        return self.ai_model_repo.find_by_id(model_id)

    # Convenience methods for model usage

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
        billing_type: str = "pay-as-you-go",
        quota_applied: bool = False,
        quota_exceeded: bool = False,
        request_id: str = "",
        endpoint: str = "",
        client_ip: Optional[str] = None,
        user_agent: Optional[str] = None,
        organization_id: Optional[str] = None
    ) -> str:
        """Log AI model usage."""
        return self.model_usage_repo.log_usage(
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

    # Convenience methods for conversations

    def get_conversation_by_session(self, session_id: str) -> Optional[Conversation]:
        """Get conversation by session ID."""
        return self.conversation_repo.find_by_session(session_id)

    def create_conversation(
        self,
        user_id: str,
        session_id: str,
        title: Optional[str] = None,
        domain: Optional[str] = None
    ) -> Conversation:
        """Create a new conversation."""
        return self.conversation_repo.create_conversation(
            user_id=user_id,
            session_id=session_id,
            title=title,
            domain=domain
        )

    def add_message_to_conversation(
        self,
        conversation_id: str,
        role: str,
        content: str,
        metadata: Optional[Dict[str, Any]] = None
    ) -> Optional[Conversation]:
        """Add a message to a conversation."""
        return self.conversation_repo.add_message_to_conversation(
            conversation_id=conversation_id,
            role=role,
            content=content,
            metadata=metadata
        )


# Create a singleton instance
_mongo_client: Optional[MongoDBClient] = None


def get_mongo_client() -> MongoDBClient:
    """Get MongoDB client singleton."""
    global _mongo_client

    if _mongo_client is None:
        _mongo_client = MongoDBClient()

    return _mongo_client
