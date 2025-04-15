from typing import Dict, List
from crewai import Agent
from langchain.tools import Tool
from Backend.ai_services.llm.llm_service import LLMService
from Backend.utils.logging_utils import get_logger
from pydantic import Field

logger = get_logger(__name__)


class ResourceAllocationAgent(Agent):
    ai_service: LLMService = Field(default_factory=LLMService)

    def __init__(self):
        # Define agent tools
        tools = [
            Tool.from_function(
                func=self.allocate_resources,
                name="allocate_resources",
                description="Optimizes resource allocation for tasks based on availability and requirements"
            ),
            Tool.from_function(
                func=self.analyze_resource_utilization,
                name="analyze_utilization",
                description="Analyzes current resource utilization and identifies optimization opportunities"
            ),
            Tool.from_function(
                func=self.forecast_resource_needs,
                name="forecast_needs",
                description="Forecasts future resource requirements based on project trends"
            ),
            Tool.from_function(
                func=self.optimize_workload_distribution,
                name="optimize_workload",
                description="Optimizes workload distribution across available resources"
            )
        ]

        super().__init__(
            role="Resource Management Specialist",
            goal="Optimize resource allocation, workload distribution, and capacity planning",
            backstory="I am an expert in resource management and allocation, using AI-driven analytics to ensure optimal distribution of resources while maintaining team efficiency and preventing burnout.",
            tools=tools,
            verbose=True
        )

    async def allocate_resources(self, task_data: Dict, available_resources: List[Dict]) -> Dict:
        """Optimize resource allocation for tasks."""
        try:
            allocation = await self.ai_service.generate_response(
                prompt="Analyze and recommend resource allocation",
                context={
                    "task": task_data,
                    "available_resources": available_resources
                }
            )

            return {
                "recommended_resources": allocation.get("recommendations", []),
                "workload_distribution": allocation.get("workload_distribution", {}),
                "efficiency_score": float(allocation.get("efficiency_score", 0.0)),
                "risk_factors": allocation.get("risk_factors", []),
                "allocation_rationale": allocation.get("allocation_rationale", {})
            }
        except Exception as e:
            logger.error(f"Resource allocation failed: {str(e)}")
            raise

    async def analyze_resource_utilization(self, resource_data: Dict) -> Dict:
        """Analyze current resource utilization patterns."""
        try:
            analysis = await self.ai_service.generate_response(
                prompt="Analyze resource utilization patterns",
                context={"resources": resource_data}
            )
            return {
                "utilization_metrics": analysis.get("utilization_metrics", {}),
                "bottlenecks": analysis.get("bottlenecks", []),
                "optimization_opportunities": analysis.get("optimization_opportunities", []),
                "efficiency_recommendations": analysis.get("recommendations", [])
            }
        except Exception as e:
            logger.error(f"Resource utilization analysis failed: {str(e)}")
            raise

    async def forecast_resource_needs(self, project_data: Dict) -> Dict:
        """Forecast future resource requirements."""
        try:
            forecast = await self.ai_service.generate_response(
                prompt="Forecast resource requirements",
                context={"project": project_data}
            )
            return {
                "resource_forecasts": forecast.get("forecasts", {}),
                "capacity_requirements": forecast.get("capacity_requirements", {}),
                "risk_assessment": forecast.get("risk_assessment", {}),
                "scaling_recommendations": forecast.get("scaling_recommendations", [])
            }
        except Exception as e:
            logger.error(f"Resource forecasting failed: {str(e)}")
            raise

    async def optimize_workload_distribution(self, workload_data: Dict) -> Dict:
        """Optimize workload distribution across resources."""
        try:
            optimization = await self.ai_service.generate_response(
                prompt="Optimize workload distribution",
                context={"workload": workload_data}
            )
            return {
                "distribution_plan": optimization.get("distribution_plan", {}),
                "load_balancing_suggestions": optimization.get("load_balancing", []),
                "efficiency_improvements": optimization.get("improvements", []),
                "implementation_steps": optimization.get("implementation_steps", [])
            }
        except Exception as e:
            logger.error(f"Workload optimization failed: {str(e)}")
            raise
