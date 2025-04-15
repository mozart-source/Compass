from typing import Dict, List
from crewai import Agent
from langchain.tools import Tool
from Backend.ai_services.task_ai.task_classification_service import TaskClassificationService
from Backend.utils.logging_utils import get_logger
from pydantic import Field
import os

logger = get_logger(__name__)

os.environ["OPENAI_API_KEY"] = "ZacAdKKUscsLMldtyUpqSJJwiaCnc5Xa"


class TaskAnalysisAgent(Agent):
    ai_service: TaskClassificationService = Field(
        default_factory=TaskClassificationService)

    def __init__(self):
        # Define agent tools
        tools = [
            Tool.from_function(
                func=self.analyze_task,
                name="analyze_task",
                description="Analyzes and classifies tasks based on their description and metadata"
            ),
            Tool.from_function(
                func=self._determine_task_type,
                name="determine_task_type",
                description="Determines the type of task based on its characteristics"
            ),
            Tool.from_function(
                func=self._extract_required_skills,
                name="extract_skills",
                description="Extracts required skills from task description"
            ),
            Tool.from_function(
                func=self._assess_risks,
                name="assess_risks",
                description="Assesses potential risks associated with the task"
            )
        ]

        super().__init__(
            role="Task Analysis Specialist",
            goal="Analyze and classify tasks accurately to optimize workflow and resource allocation",
            backstory="I am an expert in task analysis with deep understanding of project management and technical requirements. I help teams make informed decisions by providing detailed insights about tasks.",
            tools=tools,
            verbose=True
        )

    async def analyze_task(self, task_data: Dict) -> Dict:
        """Analyze a task and provide comprehensive insights."""
        try:
            # Classify task
            classification = await self.ai_service.classify_task(task_data)

            # Enhance with additional analysis
            analysis_result = {
                **classification,
                "task_type": self._determine_task_type(task_data),
                "required_skills": self._extract_required_skills(task_data),
                "risk_assessment": self._assess_risks(task_data)
            }

            return analysis_result
        except Exception as e:
            logger.error(f"Task analysis failed: {str(e)}")
            raise

    def _determine_task_type(self, task_data: Dict) -> str:
        """Determine the type of task based on its characteristics."""
        description = task_data.get('description', '').lower()
        if 'bug' in description or 'fix' in description:
            return 'bug_fix'
        elif 'feature' in description or 'implement' in description:
            return 'feature_development'
        elif 'test' in description or 'qa' in description:
            return 'testing'
        return 'general_task'

    def _extract_required_skills(self, task_data: Dict) -> List[str]:
        """Extract required skills from task description."""
        try:
            description = task_data.get('description', '').lower()
            skills = set()

            # Technical skills
            if any(tech in description for tech in ['python', 'java', 'javascript', 'react']):
                skills.add('programming')
            if 'database' in description or 'sql' in description:
                skills.add('database')
            if 'api' in description or 'rest' in description:
                skills.add('api_development')

            # Soft skills
            if 'team' in description or 'collaborate' in description:
                skills.add('teamwork')
            if 'analyze' in description or 'research' in description:
                skills.add('analytical')

            return list(skills)
        except Exception as e:
            logger.error(f"Error extracting skills: {str(e)}")
            return []

    def _assess_risks(self, task_data: Dict) -> Dict:
        """Assess potential risks associated with the task."""
        try:
            risks = []
            risk_level = "low"

            # Deadline risk
            if task_data.get('due_date') and task_data.get('estimated_hours', 0) > 20:
                risks.append("High time commitment required")
                risk_level = "medium"

            # Dependency risk
            if len(task_data.get('dependencies', [])) > 2:
                risks.append("Multiple dependencies may cause delays")
                risk_level = "high"

            # Technical risk
            if task_data.get('complexity_score', 0) > 0.7:
                risks.append("High technical complexity")
                risk_level = "high"

            return {
                "risk_level": risk_level,
                "potential_issues": risks
            }
        except Exception as e:
            logger.error(f"Error assessing risks: {str(e)}")
            return {"risk_level": "unknown", "potential_issues": []}
