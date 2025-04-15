from typing import Dict, List
from crewai import Agent
from langchain.tools import Tool
from Backend.ai_services.llm.llm_service import LLMService
from Backend.utils.logging_utils import get_logger
from pydantic import Field

logger = get_logger(__name__)


class CollaborationAgent(Agent):
    ai_service: LLMService = Field(default_factory=LLMService)

    def __init__(self):
        # Define agent tools
        tools = [
            Tool.from_function(
                func=self.analyze_collaboration,
                name="analyze_collaboration",
                description="Analyzes team collaboration patterns and suggests improvements"
            ),
            Tool.from_function(
                func=self.assess_team_dynamics,
                name="assess_team_dynamics",
                description="Assesses team dynamics and interpersonal relationships"
            ),
            Tool.from_function(
                func=self.optimize_communication,
                name="optimize_communication",
                description="Optimizes team communication channels and patterns"
            ),
            Tool.from_function(
                func=self.generate_team_recommendations,
                name="generate_recommendations",
                description="Generates personalized team improvement recommendations"
            )
        ]

        super().__init__(
            role="Team Collaboration Specialist",
            goal="Enhance team collaboration, communication efficiency, and interpersonal dynamics",
            backstory="I am an expert in team dynamics and collaboration optimization, using AI-driven insights to improve team performance and create a positive work environment.",
            tools=tools,
            verbose=True
        )

    async def analyze_collaboration(
        self,
        team_data: Dict,
        tasks: List[Dict]
    ) -> Dict:
        """Analyze team collaboration patterns and suggest improvements."""
        try:
            analysis = await self.ai_service.generate_response(
                prompt="Analyze team collaboration patterns",
                context={
                    "team_data": team_data,
                    "tasks": tasks
                }
            )

            return {
                "collaboration_score": float(analysis.get("collaboration_score", 0.0)),
                "communication_patterns": analysis.get("communication_patterns", []),
                "team_dynamics": analysis.get("team_dynamics", {}),
                "recommendations": analysis.get("recommendations", []),
                "improvement_areas": analysis.get("improvement_areas", [])
            }
        except Exception as e:
            logger.error(f"Collaboration analysis failed: {str(e)}")
            raise

    async def assess_team_dynamics(self, team_data: Dict) -> Dict:
        """Assess team dynamics and interpersonal relationships."""
        try:
            assessment = await self.ai_service.generate_response(
                prompt="Assess team dynamics and relationships",
                context={"team": team_data}
            )
            return {
                "team_cohesion": assessment.get("team_cohesion", 0.0),
                "relationship_map": assessment.get("relationship_map", {}),
                "conflict_areas": assessment.get("conflict_areas", []),
                "strength_areas": assessment.get("strength_areas", [])
            }
        except Exception as e:
            logger.error(f"Team dynamics assessment failed: {str(e)}")
            raise

    async def optimize_communication(self, team_data: Dict) -> Dict:
        """Optimize team communication channels and patterns."""
        try:
            optimization = await self.ai_service.generate_response(
                prompt="Optimize team communication",
                context={"team": team_data}
            )
            return {
                "channel_recommendations": optimization.get("channel_recommendations", {}),
                "meeting_optimizations": optimization.get("meeting_optimizations", []),
                "communication_guidelines": optimization.get("communication_guidelines", []),
                "tool_suggestions": optimization.get("tool_suggestions", [])
            }
        except Exception as e:
            logger.error(f"Communication optimization failed: {str(e)}")
            raise

    async def generate_team_recommendations(self, analysis_data: Dict) -> Dict:
        """Generate personalized team improvement recommendations."""
        try:
            recommendations = await self.ai_service.generate_response(
                prompt="Generate team improvement recommendations",
                context={"analysis": analysis_data}
            )
            return {
                "individual_recommendations": recommendations.get("individual_recommendations", {}),
                "team_level_actions": recommendations.get("team_level_actions", []),
                "leadership_suggestions": recommendations.get("leadership_suggestions", []),
                "implementation_plan": recommendations.get("implementation_plan", {})
            }
        except Exception as e:
            logger.error(f"Team recommendations generation failed: {str(e)}")
            raise
