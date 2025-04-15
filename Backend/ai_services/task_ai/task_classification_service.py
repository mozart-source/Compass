from Backend.data_layer.cache.ai_cache import cache_ai_result, get_cached_ai_result
from Backend.ai_services.embedding.embedding_service import EmbeddingService
from Backend.ai_services.base.ai_service_base import AIServiceBase
from typing import Dict, List, Optional
import aiohttp
from datetime import datetime
from Backend.core.config import settings
from Backend.utils.cache_utils import cache_response
from Backend.utils.logging_utils import get_logger
from Backend.ai_services.nlp_service.nlp_service import NLPService
from Backend.data_layer.database.models.ai_interactions import AIAgentInteraction
from Backend.data_layer.database.models.ai_models import AIModel

logger = get_logger(__name__)


class TaskClassificationService(AIServiceBase):
    def __init__(self):
        super().__init__("task_classification")
        self.embedding_service = EmbeddingService()
        self.nlp_service = NLPService()
        self.model_version = "1.0.0"

    @cache_response(ttl=3600)
    async def classify_task(
        self,
        task_data: Dict,
        db_session=None,
        user_id: Optional[int] = None,
        include_historical: bool = False
    ) -> Dict:
        """Classify task with enhanced analysis and historical comparison."""
        try:
            # Check cache
            cache_key = f"task_classification:{hash(str(task_data))}"
            if cached_result := await get_cached_ai_result(cache_key):
                return cached_result

            # Get task embedding and analysis
            task_text = f"{task_data.get('title', '')} {task_data.get('description', '')}"
            task_embedding = await self.embedding_service.get_embedding(task_text)
            sentiment = await self.nlp_service.analyze_sentiment(task_data.get('description', ''))
            entities = await self.nlp_service.extract_entities(task_text)

            # Get classification with context
            classification = await self._make_request(
                "classify",
                data={
                    "embedding": task_embedding,
                    "task_data": task_data,
                    "context": {
                        "sentiment": sentiment,
                        "entities": entities,
                        "historical_data": await self._get_historical_data(task_data) if include_historical else None
                    }
                }
            )

            result = {
                "category": classification["category"],
                "priority": classification["priority"],
                "complexity": classification["complexity"],
                "estimated_hours": classification["estimated_time"],
                "sentiment_analysis": sentiment,
                "entities": entities,
                "confidence": float(classification["confidence"]),
                "tags": classification.get("tags", []),
                "similar_tasks": classification.get("similar_tasks", [])
            }

            # Cache result
            await cache_ai_result(cache_key, result)

            # Log interaction
            if db_session and user_id is not None:
                await self._log_interaction(
                    db_session=db_session,
                    user_id=user_id,
                    input_data=task_data,
                    output_data=result,
                    success_rate=classification.get("confidence", 1.0)
                )

            return result
        except Exception as e:
            logger.error(f"Task classification error: {str(e)}")
            raise

    async def _get_historical_data(self, task_data: Dict) -> List[Dict]:
        """Get historical task data for context."""
        try:
            similar_tasks = await self._make_request(
                "find_similar_tasks",
                data={"task": task_data}
            )
            return similar_tasks.get("tasks", [])
        except Exception as e:
            logger.error(f"Error getting historical data: {str(e)}")
            return []

    async def _log_interaction(
        self,
        db_session,
        user_id: int,
        input_data: Dict,
        output_data: Dict,
        success_rate: float = 1.0
    ):
        """Log AI interaction with enhanced metrics."""
        try:
            interaction = AIAgentInteraction(
                user_id=user_id,
                ai_model_id=1,
                agent_type="task_classification",
                interaction_type="classification",
                input_data=input_data,
                output_data=output_data,
                success_rate=success_rate,
                model_version=self.model_version,
                token_usage={
                    "input_tokens": len(str(input_data)),
                    "output_tokens": len(str(output_data))
                },
                cache_hit=False,
                performance_metrics={
                    "classification_confidence": output_data.get("confidence", 0),
                    "sentiment_score": output_data.get("sentiment_analysis", {}).get("score", 0),
                    "entity_count": len(output_data.get("entities", [])),
                    "processing_time": output_data.get("processing_time", 0)
                }
            )
            db_session.add(interaction)
            await db_session.commit()
        except Exception as e:
            logger.error(f"Error logging interaction: {str(e)}")
            raise

    async def close(self):
        """Close the aiohttp session."""
        if self.session and not self.session.closed:
            await self.session.close()

    async def process_feedback(self, feedback_score: float, feedback_text: Optional[str] = None) -> None:
        """Process feedback to improve task classification model."""
        try:
            # Process feedback data
            feedback_data = {
                "feedback_score": feedback_score,
                "feedback_text": feedback_text,
                "model_version": self.model_version,
                "timestamp": datetime.utcnow().isoformat()
            }

            # Update model with feedback
            await self._make_request(
                "update_model",
                data={
                    "feedback_data": feedback_data,
                    "model_type": "task_classification"
                }
            )

            logger.info(
                f"Processed task classification feedback: {feedback_score}")

        except Exception as e:
            logger.error(
                f"Error processing task classification feedback: {str(e)}")
            # Don't raise the exception to avoid affecting the main feedback submission
            pass
