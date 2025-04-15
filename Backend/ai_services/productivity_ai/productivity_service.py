from typing import Dict, List, Optional
from datetime import datetime, timedelta
from Backend.ai_services.base.ai_service_base import AIServiceBase
from Backend.ai_services.nlp_service.nlp_service import NLPService
from Backend.utils.cache_utils import cache_response
from Backend.utils.logging_utils import get_logger
from Backend.data_layer.cache.ai_cache import cache_ai_result, get_cached_ai_result

logger = get_logger(__name__)


class ProductivityService(AIServiceBase):
    def __init__(self):
        super().__init__("productivity")
        self.nlp_service = NLPService()
        self.model_version = "1.0.0"

    @cache_response(ttl=3600)
    async def analyze_task_patterns(
        self,
        tasks: List[Dict],
        time_period: str = "daily",
        include_predictions: bool = True
    ) -> Dict:
        """Analyze task completion patterns and productivity metrics."""
        try:
            cache_key = f"task_patterns:{hash(str(tasks))}:{time_period}"
            if cached_result := await get_cached_ai_result(cache_key):
                return cached_result

            metrics = await self._calculate_task_metrics(tasks, time_period)
            insights = await self._generate_task_insights(metrics)

            result = {
                "metrics": metrics,
                "insights": insights,
                "recommendations": await self._generate_recommendations(
                    metrics["completion_rate"],
                    metrics["avg_complexity"]
                )
            }

            if include_predictions:
                result["predictions"] = await self._predict_future_metrics(metrics)

            await cache_ai_result(cache_key, result)
            return result
        except Exception as e:
            logger.error(f"Error analyzing task patterns: {str(e)}")
            raise

    async def _calculate_task_metrics(self, tasks: List[Dict], time_period: str) -> Dict:
        """Calculate comprehensive task metrics."""
        try:
            total_tasks = len(tasks)
            completed_tasks = sum(
                1 for task in tasks if task.get("status") == "completed")

            # Calculate complexity scores using NLP
            complexity_scores = []
            for task in tasks:
                if description := task.get("description"):
                    complexity = await self.nlp_service.analyze_text_complexity(description)
                    complexity_scores.append(complexity["readability_score"])

            # Time-based calculations
            time_window = self._get_time_window(time_period)
            now = datetime.utcnow()
            recent_tasks = [
                task for task in tasks
                if datetime.fromisoformat(task.get("created_at", now.isoformat())) > now - time_window
            ]

            return {
                "completion_rate": completed_tasks / total_tasks if total_tasks > 0 else 0,
                "avg_complexity": sum(complexity_scores) / len(complexity_scores) if complexity_scores else 0,
                "total_tasks": total_tasks,
                "completed_tasks": completed_tasks,
                "recent_completion_rate": self._calculate_recent_completion_rate(recent_tasks),
                "task_distribution": self._analyze_task_distribution(tasks),
                "time_metrics": self._calculate_time_metrics(tasks)
            }
        except Exception as e:
            logger.error(f"Error calculating task metrics: {str(e)}")
            raise

    async def _generate_task_insights(self, metrics: Dict) -> Dict:
        """Generate detailed insights from metrics."""
        try:
            return {
                "productivity_score": self._calculate_productivity_score(metrics),
                "efficiency_rating": self._determine_efficiency_rating(metrics),
                "trend_analysis": await self._analyze_trends(metrics),
                "bottleneck_identification": await self._identify_bottlenecks(metrics),
                "optimization_opportunities": await self._find_optimization_opportunities(metrics)
            }
        except Exception as e:
            logger.error(f"Error generating task insights: {str(e)}")
            raise

    def _get_time_window(self, time_period: str) -> timedelta:
        """Get time window for analysis."""
        return {
            "daily": timedelta(days=1),
            "weekly": timedelta(weeks=1),
            "monthly": timedelta(days=30),
            "quarterly": timedelta(days=90)
        }.get(time_period, timedelta(days=30))

    async def _predict_future_metrics(self, current_metrics: Dict) -> Dict:
        """Predict future productivity metrics."""
        try:
            return await self._make_request(
                "predict_metrics",
                data={"current_metrics": current_metrics}
            )
        except Exception as e:
            logger.error(f"Error predicting metrics: {str(e)}")
            return {}

    @cache_response(ttl=3600)
    async def analyze_workflow_efficiency(self, workflow_data: Dict) -> Dict:
        """Analyze workflow execution efficiency and bottlenecks."""
        try:
            steps = workflow_data.get("steps", [])
            total_time = workflow_data.get("actual_duration", 0)
            expected_time = workflow_data.get("estimated_duration", 0)

            efficiency_ratio = total_time / expected_time if expected_time > 0 else 0
            step_times = [step.get("duration", 0) for step in steps]
            avg_step_time = sum(step_times) / \
                len(step_times) if step_times else 0

            # Analyze step descriptions for complexity
            step_complexities = []
            for step in steps:
                description = step.get("description", "")
                if description:
                    sentiment = await self.nlp_service.analyze_sentiment(description)
                    step_complexities.append(sentiment["confidence"])

            avg_step_complexity = sum(
                step_complexities) / len(step_complexities) if step_complexities else 0

            return {
                "efficiency_metrics": {
                    "efficiency_ratio": efficiency_ratio,
                    "average_step_time": avg_step_time,
                    "total_steps": len(steps),
                    "average_step_complexity": avg_step_complexity
                },
                "optimization_suggestions": self._analyze_workflow_bottlenecks(steps)
            }
        except Exception as e:
            logger.error(f"Error analyzing workflow efficiency: {str(e)}")
            raise

    def _generate_recommendations(self, completion_rate: float, complexity: float) -> List[str]:
        """Generate productivity recommendations based on metrics."""
        recommendations = []
        if completion_rate < 0.5:
            recommendations.append(
                "Consider breaking down tasks into smaller, more manageable units")
        if complexity > 0.7:
            recommendations.append(
                "Task descriptions indicate high complexity. Consider simplifying or delegating")
        if completion_rate < 0.3 and complexity > 0.5:
            recommendations.append(
                "High task complexity may be impacting completion rates. Review task allocation")
        return recommendations

    def _analyze_workflow_bottlenecks(self, steps: List[Dict]) -> List[str]:
        """Identify workflow bottlenecks and suggest optimizations."""
        suggestions = []
        step_times = [(step.get("duration", 0), step.get(
            "name", "Unknown")) for step in steps]
        avg_time = sum(time for time, _ in step_times) / \
            len(step_times) if step_times else 0

        for duration, step_name in step_times:
            if duration > avg_time * 1.5:
                suggestions.append(
                    f"Step '{step_name}' takes significantly longer than average. Consider optimization")

        return suggestions

    def _calculate_recent_completion_rate(self, recent_tasks: List[Dict]) -> float:
        """Calculate completion rate for recent tasks."""
        if not recent_tasks:
            return 0.0
        completed = sum(1 for task in recent_tasks if task.get(
            "status") == "completed")
        return completed / len(recent_tasks)

    def _analyze_task_distribution(self, tasks: List[Dict]) -> Dict:
        """Analyze task distribution patterns."""
        categories = {}
        priorities = {}
        statuses = {}

        for task in tasks:
            category = task.get("category", "uncategorized")
            priority = task.get("priority", "medium")
            status = task.get("status", "pending")

            categories[category] = categories.get(category, 0) + 1
            priorities[priority] = priorities.get(priority, 0) + 1
            statuses[status] = statuses.get(status, 0) + 1

        return {
            "by_category": categories,
            "by_priority": priorities,
            "by_status": statuses
        }

    def _calculate_time_metrics(self, tasks: List[Dict]) -> Dict:
        """Calculate time-based metrics for tasks."""
        total_estimated = 0
        total_actual = 0
        overdue_tasks = 0
        now = datetime.utcnow()

        for task in tasks:
            estimated = task.get("estimated_hours", 0)
            actual = task.get("actual_hours", 0)
            due_date = task.get("due_date")

            total_estimated += estimated
            total_actual += actual

            if due_date and datetime.fromisoformat(due_date) < now and task.get("status") != "completed":
                overdue_tasks += 1

        return {
            "total_estimated_hours": total_estimated,
            "total_actual_hours": total_actual,
            "time_efficiency": total_estimated / total_actual if total_actual > 0 else 0,
            "overdue_tasks": overdue_tasks
        }

    def _calculate_productivity_score(self, metrics: Dict) -> float:
        """Calculate overall productivity score."""
        weights = {
            "completion_rate": 0.4,
            "time_efficiency": 0.3,
            "complexity_handling": 0.3
        }

        completion_score = metrics["completion_rate"] * 100
        time_efficiency = metrics["time_metrics"]["time_efficiency"] * \
            100 if metrics["time_metrics"]["time_efficiency"] <= 1 else 100
        complexity_score = (1 - metrics["avg_complexity"]) * 100

        return (
            weights["completion_rate"] * completion_score +
            weights["time_efficiency"] * time_efficiency +
            weights["complexity_handling"] * complexity_score
        )

    def _determine_efficiency_rating(self, metrics: Dict) -> str:
        """Determine efficiency rating based on metrics."""
        score = self._calculate_productivity_score(metrics)
        if score >= 90:
            return "excellent"
        elif score >= 75:
            return "good"
        elif score >= 60:
            return "fair"
        else:
            return "needs_improvement"

    async def _analyze_trends(self, metrics: Dict) -> Dict:
        """Analyze productivity trends."""
        try:
            return await self._make_request(
                "analyze_trends",
                data={
                    "completion_rate": metrics["completion_rate"],
                    "time_efficiency": metrics["time_metrics"]["time_efficiency"],
                    "complexity": metrics["avg_complexity"]
                }
            )
        except Exception as e:
            logger.error(f"Error analyzing trends: {str(e)}")
            return {"trend": "stable", "confidence": 0.0}

    async def _identify_bottlenecks(self, metrics: Dict) -> List[str]:
        """Identify productivity bottlenecks."""
        bottlenecks = []

        if metrics["completion_rate"] < 0.6:
            bottlenecks.append("Low task completion rate")

        if metrics["time_metrics"]["time_efficiency"] < 0.7:
            bottlenecks.append("Time management inefficiency")

        if metrics["time_metrics"]["overdue_tasks"] > 0:
            bottlenecks.append(
                f"Has {metrics['time_metrics']['overdue_tasks']} overdue tasks")

        if metrics["avg_complexity"] > 0.7:
            bottlenecks.append("High task complexity")

        return bottlenecks

    async def _find_optimization_opportunities(self, metrics: Dict) -> List[Dict]:
        """Find opportunities for productivity optimization."""
        try:
            distribution = metrics["task_distribution"]
            opportunities = []

            # Check category distribution
            if "uncategorized" in distribution["by_category"]:
                opportunities.append({
                    "type": "organization",
                    "description": "Categorize uncategorized tasks for better organization",
                    "impact": "medium"
                })

            # Check priority balance
            priorities = distribution["by_priority"]
            if priorities.get("high", 0) > len(priorities) * 0.4:
                opportunities.append({
                    "type": "prioritization",
                    "description": "Too many high-priority tasks. Consider reprioritization",
                    "impact": "high"
                })

            # Check workload distribution
            if metrics["time_metrics"]["time_efficiency"] < 0.8:
                opportunities.append({
                    "type": "time_management",
                    "description": "Improve time estimation accuracy",
                    "impact": "high"
                })

            return opportunities

        except Exception as e:
            logger.error(f"Error finding optimization opportunities: {str(e)}")
            return []
