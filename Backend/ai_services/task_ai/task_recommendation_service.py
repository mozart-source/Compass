from typing import Dict, List
import aiohttp
from Backend.core.config import settings
from Backend.utils.cache_utils import cache_response
from Backend.utils.logging_utils import get_logger
from Backend.ai_services.productivity_ai.productivity_service import ProductivityService
from Backend.ai_services.rag.rag_service import RAGService
from Backend.ai_services.embedding.embedding_service import EmbeddingService
from Backend.utils.cache_utils import cache_response
from Backend.utils.logging_utils import get_logger

logger = get_logger(__name__)

class TaskRecommendationService(AIServiceBase):
    def __init__(self):
        super().__init__("task_recommendation")
        self.rag_service = RAGService()
        self.embedding_service = EmbeddingService()

    @cache_response(ttl=1800)
    async def get_task_recommendations(
        self,
        user_id: int,
        project_id: int,
        limit: int = 5
    ) -> Dict:
        """Get personalized task recommendations."""
        try:
            # Get user context and preferences
            user_context = await self._make_request(
                "user_context",
                data={"user_id": user_id}
            )

            # Get project context
            project_context = await self._make_request(
                "project_context",
                data={"project_id": project_id}
            )

            # Get recommendations using RAG
            recommendations = await self.rag_service.query_knowledge_base(
                query=f"Recommend tasks for user {user_id} in project {project_id}",
                context={
                    "user_context": user_context,
                    "project_context": project_context
                },
                limit=limit
            )

            return {
                "recommendations": recommendations["answer"],
                "similar_tasks": recommendations["sources"],
                "confidence": recommendations["confidence"]
            }
        except Exception as e:
            logger.error(f"Task recommendation error: {str(e)}")
            raise

    @cache_response(ttl=3600)
    async def get_task_insights(self, task_id: int) -> Dict:
        """Get AI-powered insights for a specific task."""
        try:
            # Get task context from RAG
            task_context = await self.rag_service.query_knowledge_base(
                query=f"Get insights for task {task_id}",
                limit=3
            )

            # Generate insights
            insights = await self._make_request(
                "task_insights",
                data={
                    "task_id": task_id,
                    "context": task_context["answer"],
                    "similar_tasks": task_context["sources"]
                }
            )

            return {
                "insights": insights["recommendations"],
                "risk_factors": insights["risks"],
                "optimization_suggestions": insights["optimizations"],
                "confidence": float(insights.get("confidence", 0.0))
            }
        except Exception as e:
            logger.error(f"Task insights error: {str(e)}")
            raise
    def __init__(self):
        self.productivity_service = ProductivityService()
        self.api_key = settings.RECOMMENDATION_API_KEY
        self.base_url = settings.RECOMMENDATION_API_BASE_URL
        self.session = None
    async def _get_session(self) -> aiohttp.ClientSession:
        if self.session is None or self.session.closed:
            self.session = aiohttp.ClientSession(headers={
                "Authorization": f"Bearer {self.api_key}",
                "Content-Type": "application/json"
            })
        return self.session
    
    async def close(self):
        """Close the aiohttp session."""
        if self.session and not self.session.closed:
            await self.session.close()