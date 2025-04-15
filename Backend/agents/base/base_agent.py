from typing import Dict, Any, Optional
from crewai import Agent
from pydantic import Field
from Backend.utils.logging_utils import get_logger
from Backend.ai_services.base.ai_service_base import AIServiceBase

logger = get_logger(__name__)


class BaseAgent(Agent):
    ai_service: AIServiceBase = Field(...)
    name: str = Field(default="")

    def __init__(
        self,
        name: str,
        role: str,
        goal: str,
        ai_service: AIServiceBase,
        backstory: Optional[str] = None,
        verbose: bool = False
    ):
        super().__init__(
            role=role,
            goal=goal,
            backstory=backstory or f"I am an AI agent specialized in {role}",
            verbose=verbose
        )
        self.ai_service = ai_service
        self.name = name

    async def execute_task(self, task: Dict[str, Any]) -> Dict[str, Any]:
        """Execute a task using the agent's AI service."""
        try:
            logger.info(
                f"Agent {self.name} executing task: {task.get('title', '')}")
            return await self.ai_service._make_request(
                endpoint="process",
                data={"task": task}
            )
        except Exception as e:
            logger.error(f"Agent {self.name} failed to execute task: {str(e)}")
            raise
