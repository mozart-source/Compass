from typing import Dict, List
from crewai import Agent
from langchain.tools import Tool
from Backend.ai_services.workflow_ai.workflow_optimization_service import WorkflowOptimizationService
from Backend.utils.logging_utils import get_logger
from pydantic import Field

logger = get_logger(__name__)


class WorkflowOptimizationAgent(Agent):
    ai_service: WorkflowOptimizationService = Field(
        default_factory=WorkflowOptimizationService)

    def __init__(self):
        # Define agent tools
        tools = [
            Tool.from_function(
                func=self.optimize_workflow,
                name="optimize_workflow",
                description="Optimizes workflow processes and provides efficiency recommendations"
            ),
            Tool.from_function(
                func=self.analyze_workflow_patterns,
                name="analyze_patterns",
                description="Analyzes workflow patterns and identifies potential improvements"
            ),
            Tool.from_function(
                func=self.assess_workflow_risks,
                name="assess_risks",
                description="Assesses potential risks and bottlenecks in the workflow"
            ),
            Tool.from_function(
                func=self.suggest_automation,
                name="suggest_automation",
                description="Suggests potential automation opportunities in the workflow"
            )
        ]

        super().__init__(
            role="Workflow Optimization Specialist",
            goal="Optimize workflow efficiency, identify improvements, and maximize team productivity",
            backstory="I am an expert in workflow analysis and optimization, specializing in identifying inefficiencies and suggesting improvements to streamline processes and enhance team productivity.",
            tools=tools,
            verbose=True
        )

    async def optimize_workflow(self, workflow_id: int) -> Dict:
        """Optimize workflow and provide comprehensive recommendations."""
        try:
            optimization_result = await self.ai_service.optimize_workflow(workflow_id)
            patterns = await self.ai_service.analyze_workflow_patterns(workflow_id)

            return {
                **optimization_result,
                "identified_patterns": patterns["patterns"],
                "risk_assessment": patterns["risk_areas"],
                "success_probability": patterns["success_probability"],
                "optimization_score": optimization_result.get("optimization_score", 0.0),
                "suggested_improvements": optimization_result.get("improvements", [])
            }
        except Exception as e:
            logger.error(f"Workflow optimization failed: {str(e)}")
            raise

    async def analyze_workflow_patterns(self, workflow_data: Dict) -> Dict:
        """Analyze workflow patterns and identify optimization opportunities."""
        try:
            analysis = await self.ai_service.analyze_workflow_patterns(workflow_data)
            return {
                "patterns": analysis.get("patterns", []),
                "bottlenecks": analysis.get("bottlenecks", []),
                "efficiency_metrics": analysis.get("efficiency_metrics", {}),
                "improvement_areas": analysis.get("improvement_areas", [])
            }
        except Exception as e:
            logger.error(f"Workflow pattern analysis failed: {str(e)}")
            raise

    async def assess_workflow_risks(self, workflow_data: Dict) -> Dict:
        """Assess potential risks and bottlenecks in the workflow."""
        try:
            risk_assessment = await self.ai_service.assess_workflow_risks(workflow_data)
            return {
                "risk_areas": risk_assessment.get("risk_areas", []),
                "risk_scores": risk_assessment.get("risk_scores", {}),
                "mitigation_strategies": risk_assessment.get("mitigation_strategies", []),
                "impact_assessment": risk_assessment.get("impact_assessment", {})
            }
        except Exception as e:
            logger.error(f"Workflow risk assessment failed: {str(e)}")
            raise

    async def suggest_automation(self, workflow_data: Dict) -> Dict:
        """Suggest potential automation opportunities in the workflow."""
        try:
            automation_suggestions = await self.ai_service.analyze_automation_opportunities(workflow_data)
            return {
                "automation_opportunities": automation_suggestions.get("opportunities", []),
                "estimated_impact": automation_suggestions.get("impact_estimates", {}),
                "implementation_complexity": automation_suggestions.get("complexity_scores", {}),
                "roi_estimates": automation_suggestions.get("roi_estimates", {})
            }
        except Exception as e:
            logger.error(f"Automation suggestion analysis failed: {str(e)}")
            raise
