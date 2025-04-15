from typing import Dict, List
from crewai import Agent
from langchain.tools import Tool
from pydantic import Field
from Backend.ai_services.productivity_ai.productivity_service import ProductivityService
from Backend.utils.logging_utils import get_logger

logger = get_logger(__name__)


class ProductivityAgent(Agent):
    ai_service: ProductivityService = Field(
        default_factory=ProductivityService)

    def __init__(self):
        # Define agent tools
        tools = [
            Tool.from_function(
                func=self.analyze_productivity,
                name="analyze_productivity",
                description="Analyzes productivity patterns and provides optimization insights"
            ),
            Tool.from_function(
                func=self.analyze_task_efficiency,
                name="analyze_task_efficiency",
                description="Analyzes individual task efficiency and completion patterns"
            ),
            Tool.from_function(
                func=self.analyze_team_productivity,
                name="analyze_team_productivity",
                description="Analyzes team-wide productivity metrics and collaboration patterns"
            ),
            Tool.from_function(
                func=self.generate_optimization_plan,
                name="generate_optimization_plan",
                description="Generates a comprehensive productivity optimization plan"
            )
        ]

        super().__init__(
            role="Productivity Optimization Specialist",
            goal="Analyze and optimize task execution efficiency, workflow productivity, and team performance",
            backstory="I am an expert in productivity analysis and optimization, using advanced metrics and AI insights to enhance individual and team performance while maintaining sustainable work practices.",
            tools=tools,
            verbose=True
        )

    async def analyze_productivity(
        self,
        tasks: List[Dict],
        time_period: str = "daily"
    ) -> Dict:
        """Analyze productivity patterns and provide comprehensive insights."""
        try:
            patterns = await self.ai_service.analyze_task_patterns(tasks, time_period)
            workflow_data = self._aggregate_workflow_data(tasks)
            efficiency = await self.ai_service.analyze_workflow_efficiency(workflow_data)

            return {
                "task_patterns": patterns,
                "workflow_efficiency": efficiency,
                "productivity_score": patterns.get("metrics", {}).get("productivity_score", 0.0),
                "optimization_opportunities": patterns.get("insights", {}).get("optimization_opportunities", []),
                "recommendations": self._generate_productivity_recommendations(patterns, efficiency)
            }
        except Exception as e:
            logger.error(f"Productivity analysis failed: {str(e)}")
            raise

    async def analyze_task_efficiency(self, task_data: Dict) -> Dict:
        """Analyze individual task efficiency metrics."""
        try:
            efficiency_metrics = await self.ai_service.analyze_task_patterns([task_data])
            return {
                "completion_rate": efficiency_metrics.get("metrics", {}).get("completion_rate", 0.0),
                "time_efficiency": efficiency_metrics.get("metrics", {}).get("time_metrics", {}).get("time_efficiency", 0.0),
                "bottlenecks": efficiency_metrics.get("insights", {}).get("bottleneck_identification", []),
                "improvement_suggestions": efficiency_metrics.get("recommendations", [])
            }
        except Exception as e:
            logger.error(f"Task efficiency analysis failed: {str(e)}")
            raise

    async def analyze_team_productivity(self, team_data: Dict) -> Dict:
        """Analyze team-wide productivity metrics."""
        try:
            team_metrics = await self.ai_service.analyze_task_patterns(
                team_data.get("tasks", []),
                include_predictions=True
            )
            return {
                "team_efficiency": team_metrics.get("metrics", {}).get("team_efficiency", 0.0),
                "collaboration_impact": team_metrics.get("insights", {}).get("collaboration_score", 0.0),
                "workload_distribution": team_metrics.get("metrics", {}).get("task_distribution", {}),
                "improvement_areas": team_metrics.get("insights", {}).get("improvement_areas", [])
            }
        except Exception as e:
            logger.error(f"Team productivity analysis failed: {str(e)}")
            raise

    async def generate_optimization_plan(self, analysis_data: Dict) -> Dict:
        """Generate a comprehensive productivity optimization plan."""
        try:
            optimization_suggestions = await self.ai_service.analyze_task_patterns(
                analysis_data.get("tasks", []),
                include_predictions=True
            )
            return {
                "short_term_actions": optimization_suggestions.get("recommendations", []),
                "long_term_strategies": optimization_suggestions.get("insights", {}).get("optimization_opportunities", []),
                "expected_impact": optimization_suggestions.get("predictions", {}),
                "implementation_timeline": self._generate_implementation_timeline(optimization_suggestions)
            }
        except Exception as e:
            logger.error(f"Optimization plan generation failed: {str(e)}")
            raise

    def _aggregate_workflow_data(self, tasks: List[Dict]) -> Dict:
        """Aggregate tasks into workflow data."""
        return {
            "steps": [self._convert_task_to_step(task) for task in tasks],
            "estimated_duration": sum(task.get("estimated_hours", 0) for task in tasks),
            "actual_duration": sum(task.get("actual_hours", 0) for task in tasks)
        }

    def _convert_task_to_step(self, task: Dict) -> Dict:
        """Convert task data to workflow step format."""
        return {
            "name": task.get("title", "Unnamed Task"),
            "description": task.get("description", ""),
            "duration": task.get("actual_hours", 0),
            "status": task.get("status", "pending")
        }

    def _generate_productivity_recommendations(self, patterns: Dict, efficiency: Dict) -> List[str]:
        """Generate detailed productivity recommendations."""
        recommendations = []
        metrics = patterns.get("metrics", {})

        if metrics.get("completion_rate", 0) < 0.7:
            recommendations.append(
                "Improve task completion rate through better task breakdown and prioritization")

        if efficiency.get("efficiency_metrics", {}).get("efficiency_ratio", 0) < 0.8:
            recommendations.append(
                "Optimize workflow processes to reduce time inefficiencies")

        if metrics.get("avg_complexity", 0) > 0.7:
            recommendations.append(
                "Consider simplifying complex tasks or providing additional resources")

        return recommendations

    def _generate_implementation_timeline(self, optimization_data: Dict) -> Dict:
        """Generate implementation timeline for optimization suggestions."""
        timeline = {
            "immediate": [],
            "short_term": [],
            "long_term": []
        }

        for suggestion in optimization_data.get("recommendations", []):
            if "urgent" in suggestion.lower():
                timeline["immediate"].append(suggestion)
            elif "complex" in suggestion.lower():
                timeline["long_term"].append(suggestion)
            else:
                timeline["short_term"].append(suggestion)

        return timeline
