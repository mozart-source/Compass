from typing import Dict, List, Optional, Any, cast, Sequence, Union
from crewai import Crew, Agent, Task
from langchain.tools import Tool
from agents.task_agents.task_analysis_agent import TaskAnalysisAgent
from agents.task_agents.task_management_agent import TaskManagementAgent
from agents.workflow_agents.workflow_optimization_agent import WorkflowOptimizationAgent
from agents.productivity_agents.productivity_agent import ProductivityAgent
from agents.collaboration_agents.collaboration_agent import CollaborationAgent
from agents.resource_agents.resource_allocation_agent import ResourceAllocationAgent
from orchestration.compass_task import CompassTask
from utils.logging_utils import get_logger
from data_layer.repositories.task_repository import TaskRepository
from data_layer.database.connection import get_db
from core.config import settings

from datetime import datetime

logger = get_logger(__name__)


class CrewOrchestrator:
    def __init__(self):
        # Configure OpenAI client for DeepInfra
        import os
        from openai import OpenAI

        os.environ["OPENAI_API_KEY"] = settings.LLM_API_KEY
        os.environ["OPENAI_API_BASE"] = settings.LLM_API_BASE_URL

        self.client = OpenAI(
            api_key=settings.LLM_API_KEY,
            base_url=settings.LLM_API_BASE_URL
        )
        self.task_analyzer: Agent = TaskAnalysisAgent()
        self.task_manager: Agent = TaskManagementAgent()
        self.workflow_optimizer: Agent = WorkflowOptimizationAgent()
        self.productivity_agent: Agent = ProductivityAgent()
        self.collaboration_agent: Agent = CollaborationAgent()
        self.resource_agent: Agent = ResourceAllocationAgent()

    async def process_task(self, task_data: Dict[str, Any], team_data: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """Orchestrate task processing using CrewAI framework."""
        try:
            # Define tasks for the crew
            tasks: List[Task] = []

            # Create analysis task
            analysis_task = CompassTask(
                description="Analyze and classify the task",
                agent=self.task_analyzer,
                expected_output="Detailed task analysis and classification",
                task_data=task_data,
                task_type="analysis"
            )
            tasks.append(analysis_task)

            # Create planning task
            planning_task = CompassTask(
                description="Create and plan the task",
                agent=self.task_manager,
                expected_output="Task creation and planning details",
                task_data=task_data,
                task_type="creation"
            )
            tasks.append(planning_task)

            # Create resource allocation task
            resource_task = CompassTask(
                description="Allocate resources for the task",
                agent=self.resource_agent,
                expected_output="Resource allocation plan",
                task_data=task_data,
                task_type="resource_allocation"
            )
            tasks.append(resource_task)

            if workflow_id := task_data.get('workflow_id'):
                workflow_task = CompassTask(
                    description="Optimize workflow for the task",
                    agent=self.workflow_optimizer,
                    expected_output="Workflow optimization recommendations",
                    task_data={"workflow_id": workflow_id,
                               "task_data": task_data},
                    task_type="workflow_optimization"
                )
                tasks.append(workflow_task)

            if team_data:
                collab_task = CompassTask(
                    description="Analyze team collaboration impact",
                    agent=self.collaboration_agent,
                    expected_output="Collaboration analysis and recommendations",
                    task_data={"team_data": team_data,
                               "task_data": task_data},
                    task_type="collaboration_analysis"
                )
                tasks.append(collab_task)

            # Create and execute the crew
            crew = Crew(
                agents=[task.agent for task in tasks if task.agent is not None],
                tasks=cast(List[Task], tasks),  # Cast to help type checker
                verbose=True
            )

            # Execute crew tasks
            try:
                results = crew.kickoff()
                return self._process_crew_results(results)
            except Exception as e:
                logger.error(f"Crew execution failed: {str(e)}")
                return {"error": f"Crew execution failed: {str(e)}"}

        except Exception as e:
            logger.error(f"Task processing failed: {str(e)}")
            raise

    def _process_crew_results(self, results: Union[str, Dict[str, Any], Any]) -> Dict[str, Any]:
        """Process and structure crew execution results."""
        try:
            if isinstance(results, str):
                return {"output": results}

            if not results:
                return {"error": "No results from crew execution"}

            if isinstance(results, dict):
                processed_results = {}
                for task_name, result in results.items():
                    if isinstance(result, dict):
                        processed_results[task_name.lower().replace(" ", "_")] = {
                            "output": result.get("output"),
                            "status": result.get("status", "completed"),
                            "error": result.get("error")
                        }
                    else:
                        processed_results[task_name.lower().replace(" ", "_")] = {
                            "output": str(result),
                            "status": "completed"
                        }
                return processed_results

            return {"output": str(results)}
        except Exception as e:
            logger.error(f"Error processing crew results: {str(e)}")
            return {"error": str(e)}

    async def update_task(self, task_id: str, updates: Dict, current_state: Dict) -> Dict[str, Any]:
        """Coordinate task updates using CrewAI framework."""
        try:
            tasks: List[Task] = []

            # Create update task
            update_task = CompassTask(
                description=f"Update task {task_id} with new information",
                agent=self.task_manager,
                expected_output="Updated task details",
                task_data={
                    "task_id": task_id,
                    "updates": updates,
                    "current_state": current_state
                },
                task_type="update"
            )
            tasks.append(update_task)

            if 'team_impact' in updates:
                impact_task = CompassTask(
                    description="Analyze team impact of task updates",
                    agent=self.collaboration_agent,
                    expected_output="Team impact analysis",
                    task_data={
                        "task_id": task_id,
                        "updates": updates
                    },
                    task_type="impact_analysis"
                )
                tasks.append(impact_task)

            if updates.get('workflow_changes'):
                workflow_task = CompassTask(
                    description="Re-optimize workflow after task updates",
                    agent=self.workflow_optimizer,
                    expected_output="Updated workflow optimization",
                    task_data={
                        "task_id": task_id,
                        "updates": updates
                    },
                    task_type="workflow_update"
                )
                tasks.append(workflow_task)

            crew = Crew(
                agents=[task.agent for task in tasks if task.agent is not None],
                tasks=cast(List[Task], tasks),  # Cast to help type checker
                verbose=True
            )

            try:
                results = crew.kickoff()
                return self._process_crew_results(results)
            except Exception as e:
                logger.error(f"Crew execution failed: {str(e)}")
                return {"error": f"Crew execution failed: {str(e)}"}

        except Exception as e:
            logger.error(f"Task update failed: {str(e)}")
            raise

    async def delete_task(self, task_id: str, task_data: Dict) -> Dict[str, Any]:
        """Coordinate task deletion using CrewAI framework."""
        try:
            tasks: List[Task] = []

            # Create deletion analysis task
            deletion_task = CompassTask(
                description=f"Analyze impact of deleting task {task_id}",
                agent=self.task_manager,
                expected_output="Task deletion impact analysis",
                task_data={
                    "task_id": task_id,
                    "task_data": task_data
                },
                task_type="deletion_analysis"
            )
            tasks.append(deletion_task)

            if workflow_id := task_data.get('workflow_id'):
                workflow_task = CompassTask(
                    description="Assess workflow impact of task deletion",
                    agent=self.workflow_optimizer,
                    expected_output="Workflow impact assessment",
                    task_data={
                        "task_id": task_id,
                        "workflow_id": workflow_id
                    },
                    task_type="workflow_impact"
                )
                tasks.append(workflow_task)

            crew = Crew(
                agents=[task.agent for task in tasks if task.agent is not None],
                tasks=cast(List[Task], tasks),  # Cast to help type checker
                verbose=True
            )

            try:
                results = crew.kickoff()
                return self._process_crew_results(results)
            except Exception as e:
                logger.error(f"Crew execution failed: {str(e)}")
                return {"error": f"Crew execution failed: {str(e)}"}

        except Exception as e:
            logger.error(f"Task deletion failed: {str(e)}")
            raise

    async def process_db_task(self, db_task_id: int) -> Dict[str, Any]:
        """Process a task directly from the database.

        Args:
            db_task_id: Database task ID

        Returns:
            Processing results
        """
        try:
            # Get database session
            async for db in get_db():
                # Get task from database
                repo = TaskRepository(db)
                db_task = await repo.get_task_with_details(db_task_id)

                if not db_task:
                    return {"error": f"Task with ID {db_task_id} not found"}

                # Convert to CompassTask
                compass_task = CompassTask.from_db_task(
                    db_task=db_task,
                    agent=self.task_manager,
                    expected_output="Task processing results"
                )

                # Create crew with single task
                crew = Crew(
                    agents=[compass_task.agent] if compass_task.agent else [
                        self.task_manager],
                    # Cast to help type checker
                    tasks=[cast(Task, compass_task)],
                    verbose=True
                )

                # Execute crew
                results = crew.kickoff()
                processed_results = self._process_crew_results(results)

                # Update task in database with AI results if available
                if "output" in processed_results and not processed_results.get("error"):
                    task_updates = compass_task.to_db_task_update()
                    if task_updates:
                        # Add AI suggestions
                        if "ai_suggestions" not in task_updates:
                            task_updates["ai_suggestions"] = {}

                        task_updates["ai_suggestions"]["crew_processing"] = {
                            "timestamp": str(datetime.now()),
                            "result": processed_results.get("output")
                        }

                        # Update task in database
                        await repo.update_task(db_task_id, task_updates)

                return processed_results

            # Return error if we couldn't get a database session
            return {"error": "Could not establish database connection"}

        except Exception as e:
            logger.error(f"Database task processing failed: {str(e)}")
            return {"error": f"Database task processing failed: {str(e)}"}

    async def close(self):
        """Close all resources used by the orchestrator."""
        logger.info("Closing CrewOrchestrator resources")
        # No actual resources to close in this implementation
        # This method exists to maintain compatibility with other services
        pass
