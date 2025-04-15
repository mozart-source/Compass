from typing import Dict, List, Optional
from Backend.ai_services.base.ai_service_base import AIServiceBase
from Backend.ai_services.nlp_service.nlp_service import NLPService
from Backend.utils.cache_utils import cache_response
from Backend.utils.logging_utils import get_logger
from Backend.data_layer.cache.ai_cache import cache_ai_result, get_cached_ai_result

logger = get_logger(__name__)

class SummarizationService(AIServiceBase):
    def __init__(self):
        super().__init__("summarization")
        self.nlp_service = NLPService()
        self.model_version = "1.0.0"
        self.supported_languages = ["en", "es", "fr", "de"]

    @cache_response(ttl=7200)
    async def generate_summary(
        self,
        text: str,
        max_length: int = 130,
        min_length: int = 30,
        language: str = "en",
        format_template: Optional[str] = None
    ) -> Dict:
        """Generate a concise summary with enhanced features."""
        try:
            cache_key = f"summary:{hash(text)}:{max_length}:{language}"
            if cached_result := await get_cached_ai_result(cache_key):
                return cached_result

            if language not in self.supported_languages:
                language = "en"

            result = await self._make_request(
                "summarize",
                data={
                    "text": text,
                    "max_length": max_length,
                    "min_length": min_length,
                    "language": language,
                    "template": format_template
                }
            )

            keywords = await self.nlp_service.extract_keywords(text)
            entities = await self.nlp_service.extract_entities(text)
            
            summary_result = {
                "summary": result["summary"],
                "key_points": keywords,
                "entities": entities,
                "original_length": len(text.split()),
                "summary_length": len(result["summary"].split()),
                "language": language,
                "metrics": {
                    "compression_ratio": len(result["summary"].split()) / len(text.split()),
                    "keyword_density": len(keywords) / len(text.split()),
                    "readability_score": result.get("readability_score", 0.0)
                }
            }

            await cache_ai_result(cache_key, summary_result)
            return summary_result

        except Exception as e:
            logger.error(f"Summary generation error: {str(e)}")
            raise

    @cache_response(ttl=7200)
    async def summarize_workflow(
        self,
        workflow_data: Dict,
        include_metrics: bool = True
    ) -> Dict:
        """Generate workflow summary with comprehensive metrics."""
        try:
            steps_text = " ".join([
                f"{step.get('description', '')} {step.get('notes', '')}"
                for step in workflow_data.get("steps", [])
            ])
            
            summary = await self.generate_summary(
                steps_text,
                max_length=200,
                format_template="workflow"
            )
            
            result = {
                "workflow_summary": summary["summary"],
                "key_steps": summary["key_points"],
                "entities": summary["entities"],
                "metrics": {
                    "total_steps": len(workflow_data.get("steps", [])),
                    "compression_ratio": summary["metrics"]["compression_ratio"],
                    "complexity_score": await self._calculate_workflow_complexity(workflow_data)
                } if include_metrics else None
            }

            return result

        except Exception as e:
            logger.error(f"Workflow summarization error: {str(e)}")
            raise

    async def _calculate_workflow_complexity(self, workflow_data: Dict) -> float:
        """Calculate workflow complexity score."""
        try:
            complexity_factors = {
                "step_count": len(workflow_data.get("steps", [])),
                "dependencies": len(workflow_data.get("dependencies", [])),
                "branching_factor": len(workflow_data.get("conditions", [])),
                "estimated_duration": workflow_data.get("estimated_duration", 0)
            }
            
            return await self._make_request(
                "calculate_complexity",
                data=complexity_factors
            )
        except Exception as e:
            logger.error(f"Complexity calculation error: {str(e)}")
            return 0.0

    @cache_response(ttl=7200)
    async def summarize_task_group(self, tasks: List[Dict]) -> Dict:
        """Generate a summary for a group of related tasks."""
        try:
            tasks_text = " ".join([f"{task.get('title', '')} {task.get('description', '')}" for task in tasks])
            summary = await self.generate_summary(tasks_text)
            
            return {
                "group_summary": summary["summary"],
                "common_themes": summary["key_points"],
                "task_count": len(tasks),
                "summary_metrics": {
                    "original_length": summary["original_length"],
                    "summary_length": summary["summary_length"]
                }
            }
        except Exception as e:
            logger.error(f"Error summarizing task group: {str(e)}")
            raise

    async def close(self):
        """Close the aiohttp session."""
        if self.session and not self.session.closed:
            await self.session.close()