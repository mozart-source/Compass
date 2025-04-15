from Backend.celery_app import celery_app
from Backend.ai_services.llm.llm_service import LLMService
from Backend.ai_services.task_ai.task_classification_service import TaskClassificationService
from typing import Dict, Optional
from typing import List
from celery import shared_task
from Backend.orchestration.crew_orchestrator import CrewOrchestrator
from Backend.utils.logging_utils import get_logger
from datetime import datetime

logger = get_logger(__name__)

@shared_task(name="tasks.ai_tasks.process_task_analysis")
async def process_task_analysis(task_data: dict) -> dict:
    """Process task analysis using AI agents."""
    try:
        orchestrator = CrewOrchestrator()
        result = await orchestrator.analyze_and_optimize_task(task_data)
        return result
    except Exception as e:
        logger.error(f"Task analysis failed: {str(e)}")
        raise

@shared_task(name="tasks.ai_tasks.generate_productivity_insights")
async def generate_productivity_insights(user_id: int, interval: str, metrics: List[str]) -> dict:
    """Generate AI-powered productivity insights."""
    try:
        orchestrator = CrewOrchestrator()
        insights = await orchestrator.analyze_productivity(user_id, interval, metrics)
        return insights
    except Exception as e:
        logger.error(f"Productivity analysis failed: {str(e)}")
        raise
@celery_app.task(
    name="tasks.ai_tasks.classify_task_async",
    queue="ai",
    priority=5,
    rate_limit="50/m"
)
async def classify_task_async(
    task_data: Dict,
    user_id: int,
    session = None
) -> Dict:
    """Process task classification asynchronously."""
    try:
        classifier = TaskClassificationService()
        result = await classifier.classify_task(
            task_data=task_data,
            db_session=session
        )
        return {
            "status": "success",
            "classification": result,
            "user_id": user_id
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "user_id": user_id
        }

@celery_app.task(
    name="tasks.ai_tasks.process_llm_request",
    queue="ai",
    priority=7,
    rate_limit="30/m"
)
async def process_llm_request(
    prompt: str,
    user_id: int,
    context: Optional[Dict] = None,
    model_params: Optional[Dict] = None
) -> Dict:
    """Process LLM requests asynchronously."""
    try:
        llm_service = LLMService()
        result = await llm_service.generate_response(
            prompt=prompt,
            context=context,
            model_parameters=model_params
        )
        return {
            "status": "success",
            "response": result,
            "user_id": user_id,
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "timestamp": datetime.utcnow().isoformat()
        }

@celery_app.task(
    name="tasks.ai_tasks.batch_process_tasks",
    queue="ai",
    priority=4
)
async def batch_process_tasks(
    tasks: List[Dict],
    process_type: str,
    user_id: int
) -> Dict:
    """Process multiple tasks in batch using AI."""
    try:
        llm_service = LLMService()
        classifier = TaskClassificationService()
        
        results = []
        for task in tasks:
            if process_type == "classification":
                result = await classifier.classify_task(task)
            elif process_type == "enhancement":
                result = await llm_service.enhance_task_description(task)
            else:
                raise ValueError(f"Unknown process type: {process_type}")
            
            results.append({
                "task_id": task.get("id"),
                "result": result
            })

        return {
            "status": "success",
            "results": results,
            "processed_count": len(results),
            "user_id": user_id,
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "timestamp": datetime.utcnow().isoformat()
        }

@celery_app.task(
    name="tasks.ai_tasks.analyze_workflow_efficiency",
    queue="ai",
    priority=6
)
async def analyze_workflow_efficiency(
    workflow_id: int,
    historical_data: List[Dict],
    user_id: int
) -> Dict:
    """Analyze workflow efficiency using AI."""
    try:
        llm_service = LLMService()
        
        # Analyze workflow patterns
        analysis = await llm_service.analyze_workflow(
            workflow_id=workflow_id,
            historical_data=historical_data
        )
        
        return {
            "status": "success",
            "workflow_id": workflow_id,
            "efficiency_score": analysis.get("efficiency_score"),
            "bottlenecks": analysis.get("bottlenecks", []),
            "recommendations": analysis.get("recommendations", []),
            "user_id": user_id,
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "timestamp": datetime.utcnow().isoformat()
        }

@celery_app.task(
    name="tasks.ai_tasks.process_document",
    queue="ai",
    priority=5,
    rate_limit="20/m"
)
async def process_document(
    document_data: Dict,
    user_id: int,
    processing_type: str
) -> Dict:
    """Process documents using AI for various purposes."""
    try:
        llm_service = LLMService()
        result = await llm_service.process_document(
            content=document_data.get("content"),
            doc_type=document_data.get("type"),
            processing_type=processing_type
        )
        
        return {
            "status": "success",
            "document_id": document_data.get("id"),
            "processing_type": processing_type,
            "result": result,
            "user_id": user_id,
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "timestamp": datetime.utcnow().isoformat()
        }

@celery_app.task(
    name="tasks.ai_tasks.generate_meeting_summary",
    queue="ai",
    priority=6
)
async def generate_meeting_summary(
    meeting_data: Dict,
    user_id: int
) -> Dict:
    """Generate AI-powered meeting summaries."""
    try:
        llm_service = LLMService()
        summary = await llm_service.summarize_meeting(
            transcript=meeting_data.get("transcript"),
            participants=meeting_data.get("participants"),
            duration=meeting_data.get("duration")
        )
        
        return {
            "status": "success",
            "meeting_id": meeting_data.get("id"),
            "summary": summary.get("summary"),
            "action_items": summary.get("action_items", []),
            "key_points": summary.get("key_points", []),
            "user_id": user_id,
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "timestamp": datetime.utcnow().isoformat()
        }

@celery_app.task(
    name="tasks.ai_tasks.recommend_tasks",
    queue="ai",
    priority=4
)
async def recommend_tasks(
    user_id: int,
    user_context: Dict,
    max_recommendations: int = 5
) -> Dict:
    """Generate personalized task recommendations."""
    try:
        llm_service = LLMService()
        recommendations = await llm_service.generate_task_recommendations(
            user_context=user_context,
            limit=max_recommendations
        )
        
        return {
            "status": "success",
            "user_id": user_id,
            "recommendations": recommendations,
            "recommendation_count": len(recommendations),
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "timestamp": datetime.utcnow().isoformat()
        }

@celery_app.task(
    name="tasks.ai_tasks.process_text_analysis",
    queue="ai",
    priority=5,
    rate_limit="30/m"
)
async def process_text_analysis(
    text: str,
    analysis_type: str,
    user_id: int,
    options: Optional[Dict] = None
) -> Dict:
    """Process text analysis using AI services."""
    try:
        llm_service = LLMService()
        result = await llm_service.analyze_text(
            text=text,
            analysis_type=analysis_type,
            options=options
        )
        
        return {
            "status": "success",
            "result": result,
            "user_id": user_id,
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "timestamp": datetime.utcnow().isoformat()
        }
    """Generate personalized task recommendations."""
    try:
        llm_service = LLMService()
        recommendations = await llm_service.generate_task_recommendations(
            user_context=user_context,
            limit=max_recommendations
        )
        
        return {
            "status": "success",
            "user_id": user_id,
            "recommendations": recommendations,
            "recommendation_count": len(recommendations),
            "timestamp": datetime.utcnow().isoformat()
        }
    except Exception as e:
        return {
            "status": "error",
            "error": str(e),
            "timestamp": datetime.utcnow().isoformat()
        }
