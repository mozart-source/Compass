from typing import Dict, List, Optional
from Backend.ai_services.base.ai_service_base import AIServiceBase
from Backend.ai_services.rag.rag_service import RAGService
from Backend.ai_services.summarization_engine.summarization_service import SummarizationService
from Backend.utils.cache_utils import cache_response
from Backend.utils.logging_utils import get_logger
from Backend.data_layer.cache.ai_cache import cache_ai_result, get_cached_ai_result
import datetime

logger = get_logger(__name__)

class WorkflowOptimizationService(AIServiceBase):
    def __init__(self):
        super().__init__("workflow_optimization")
        self.rag_service = RAGService()
        self.summarization_service = SummarizationService()
        self.model_version = "1.0.0"

    @cache_response(ttl=7200)
    async def optimize_workflow(self, workflow_id: int, include_historical: bool = True) -> Dict:
        """Generate workflow optimization recommendations."""
        try:
            cache_key = f"workflow_opt:{workflow_id}"
            if cached_result := await get_cached_ai_result(cache_key):
                return cached_result

            # Get workflow context and historical analysis
            workflow_context = await self.rag_service.query_knowledge_base(
                query=f"Analyze workflow {workflow_id} efficiency",
                limit=5,
                filters={"workflow_id": workflow_id}
            )

            # Get workflow summary
            workflow_summary = await self.summarization_service.summarize_workflow(
                workflow_context["sources"][0]
            )

            # Generate optimization recommendations
            optimization = await self._make_request(
                "optimize",
                data={
                    "workflow_id": workflow_id,
                    "context": workflow_context["answer"],
                    "historical_data": workflow_context["sources"],
                    "summary": workflow_summary,
                    "include_historical": include_historical
                }
            )
            async def process_feedback(self, feedback_score: float, feedback_text: Optional[str] = None) -> None:
                """Process feedback to improve workflow optimization recommendations."""
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
                            "model_type": "workflow_optimization"
                        }
                    )
                    
                    logger.info(f"Processed workflow optimization feedback: {feedback_score}")
                    
                except Exception as e:
                        logger.error(f"Error processing workflow optimization feedback: {str(e)}")
                        # Don't raise the exception to avoid affecting the main feedback submission
                        pass

            result = {
                "recommendations": optimization["recommendations"],
                "bottlenecks": optimization["bottlenecks"],
                "efficiency_score": float(optimization.get("efficiency_score", 0.0)),
                "estimated_improvement": float(optimization.get("improvement", 0.0)),
                "workflow_summary": workflow_summary,
                "optimization_metrics": {
                    "confidence": float(optimization.get("confidence", 0.0)),
                    "impact_score": float(optimization.get("impact_score", 0.0)),
                    "risk_level": optimization.get("risk_level", "low")
                }
            }

            await cache_ai_result(cache_key, result)
            return result

        except Exception as e:
            logger.error(f"Workflow optimization error: {str(e)}")
            raise

    async def analyze_workflow_patterns(
        self,
        workflow_id: int,
        time_range: Optional[str] = "1month"
    ) -> Dict:
        """Analyze workflow patterns and suggest improvements."""
        try:
            patterns = await self._make_request(
                "analyze_patterns",
                data={
                    "workflow_id": workflow_id,
                    "time_range": time_range
                }
            )

            historical_analysis = await self._analyze_historical_patterns(workflow_id)

            return {
                "patterns": patterns["identified_patterns"],
                "suggestions": patterns["improvement_suggestions"],
                "risk_areas": patterns["risk_areas"],
                "success_probability": float(patterns.get("success_probability", 0.0)),
                "historical_trends": historical_analysis["trends"],
                "pattern_metrics": {
                    "pattern_confidence": float(patterns.get("pattern_confidence", 0.0)),
                    "pattern_frequency": patterns.get("pattern_frequency", {}),
                    "impact_analysis": patterns.get("impact_analysis", {})
                }
            }
        except Exception as e:
            logger.error(f"Workflow pattern analysis error: {str(e)}")
            raise

    async def _analyze_historical_patterns(self, workflow_id: int) -> Dict:
        """Analyze historical workflow patterns."""
        try:
            historical_data = await self._make_request(
                "historical_patterns",
                data={"workflow_id": workflow_id}
            )
            return {
                "trends": historical_data.get("trends", []),
                "performance_metrics": historical_data.get("performance_metrics", {}),
                "optimization_history": historical_data.get("optimization_history", [])
            }
        except Exception as e:
            logger.error(f"Historical pattern analysis error: {str(e)}")
            return {"trends": [], "performance_metrics": {}, "optimization_history": []}