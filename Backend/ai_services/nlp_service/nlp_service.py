from typing import Dict, List, Optional
from Backend.ai_services.base.ai_service_base import AIServiceBase
from Backend.utils.cache_utils import cache_response
from Backend.data_layer.cache.ai_cache import cache_ai_result, get_cached_ai_result
from Backend.utils.logging_utils import get_logger

logger = get_logger(__name__)

class NLPService(AIServiceBase):
    def __init__(self):
        super().__init__("nlp")
        self.model_version = "1.0.0"
        self.supported_languages = ["en", "es", "fr", "de", "ar"]

    @cache_response(ttl=3600)
    async def analyze_sentiment(self, text: str, language: str = "en") -> Dict:
        """Analyze the sentiment of given text."""
        try:
            cache_key = f"sentiment:{hash(text)}:{language}"
            if cached_result := await get_cached_ai_result(cache_key):
                return cached_result

            result = await self._make_request(
                "sentiment",
                data={
                    "text": text,
                    "language": language if language in self.supported_languages else "en"
                }
            )
            await cache_ai_result(cache_key, result)
            return result
        except Exception as e:
            logger.error(f"Sentiment analysis error: {str(e)}")
            raise

    @cache_response(ttl=3600)
    async def classify_text(
        self,
        text: str,
        labels: Optional[List[str]] = None,
        threshold: float = 0.5
    ) -> Dict:
        """Classify text into predefined categories."""
        try:
            payload = {
                "text": text,
                "threshold": threshold
            }
            if labels:
                payload["labels"] = labels

            return await self._make_request("classify", data=payload)
        except Exception as e:
            logger.error(f"Text classification error: {str(e)}")
            raise

    @cache_response(ttl=3600)
    async def extract_entities(
        self,
        text: str,
        entity_types: Optional[List[str]] = None
    ) -> List[Dict]:
        """Extract named entities from text."""
        try:
            result = await self._make_request(
                "entities",
                data={
                    "text": text,
                    "entity_types": entity_types or ["PERSON", "ORG", "DATE", "TECH"]
                }
            )
            return result.get("entities", [])
        except Exception as e:
            logger.error(f"Entity extraction error: {str(e)}")
            return []

    @cache_response(ttl=3600)
    async def extract_keywords(
        self,
        text: str,
        top_k: int = 5,
        min_score: float = 0.3
    ) -> List[Dict]:
        """Extract key phrases with relevance scores."""
        try:
            response = await self._make_request(
                "keywords",
                data={
                    "text": text,
                    "top_k": top_k,
                    "min_score": min_score
                }
            )
            return [
                {"keyword": k, "score": s}
                for k, s in response.get("keywords", {}).items()
                if s >= min_score
            ]
        except Exception as e:
            logger.error(f"Keyword extraction error: {str(e)}")
            return []

    async def analyze_text_complexity(self, text: str) -> Dict:
        """Analyze text complexity metrics."""
        try:
            return await self._make_request(
                "complexity",
                data={"text": text}
            )
        except Exception as e:
            logger.error(f"Complexity analysis error: {str(e)}")
            return {
                "readability_score": 0.0,
                "complexity_level": "medium",
                "technical_terms": []
            }

    async def process_feedback(self, feedback_score: float, feedback_text: Optional[str] = None):
        """Process feedback to improve NLP service performance."""
        try:
            # Log feedback for model improvement
            feedback_data = {
                "service": "nlp",
                "model_version": self.model_version,
                "feedback_score": feedback_score,
                "feedback_text": feedback_text,
                "timestamp": datetime.utcnow().isoformat()
            }
            
            # Store feedback for model retraining
            await self._make_request(
                "store_feedback",
                data=feedback_data
            )
            
            logger.info(f"Processed NLP service feedback: {feedback_score}")
        except Exception as e:
            logger.error(f"Failed to process NLP feedback: {str(e)}")
            # Don't raise to avoid affecting the main flow
            pass