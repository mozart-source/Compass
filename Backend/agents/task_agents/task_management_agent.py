from typing import Dict, List, Optional, Any
from datetime import datetime
from crewai import Agent
from langchain.tools import Tool
from Backend.ai_services.llm.llm_service import LLMService
from Backend.utils.logging_utils import get_logger
from pydantic import Field

logger = get_logger(__name__)


class TaskManagementAgent(Agent):
    ai_service: LLMService = Field(default_factory=LLMService)

    def __init__(self):
        # Define agent tools
        tools = [
            Tool.from_function(
                func=self.create_task,
                name="create_task",
                description="Creates a new task with AI-enhanced metadata and planning"
            ),
            Tool.from_function(
                func=self.update_task,
                name="update_task",
                description="Updates existing tasks with impact analysis and recommendations"
            ),
            Tool.from_function(
                func=self.plan_task_timeline,
                name="plan_task_timeline",
                description="Plans optimal task timeline considering team capacity and dependencies"
            )
        ]

        super().__init__(
            role="Task Management Specialist",
            goal="Optimize task lifecycle management and resource planning",
            backstory="I am an expert in task management and planning, using AI to enhance decision-making and ensure optimal workflow execution.",
            tools=tools,
            verbose=True,
            allow_delegation=True,
            memory=True
        )

    async def create_task(
        self,
        title: str,
        description: str,
        priority: Optional[str] = "medium",
        due_date: Optional[str] = None,
        assignee: Optional[Dict] = None
    ) -> Dict:
        """Create a new task with AI-enhanced metadata."""
        try:
            task_data = {
                "title": title,
                "description": description,
                "priority": priority,
                "due_date": due_date,
                "assignee": assignee,
                "status": "pending",
                "created_at": datetime.utcnow().isoformat()
            }

            # Get AI recommendations for task
            enhancement = await self.ai_service.generate_response(
                prompt=f"""Analyze this task and provide recommendations:
Title: {title}
Description: {description}
Priority: {priority}
Due Date: {due_date}

Please provide:
1. Estimated hours to complete
2. Required skills
3. Recommended priority level
4. Suggested deadline
5. Potential blockers""",
                context={
                    "system_message": "You are a task management AI assistant. Analyze tasks and provide structured recommendations. Be specific and practical in your suggestions."
                }
            )

            # Parse the response
            try:
                response_text = enhancement.get("text", "")
                # Extract numeric values and lists from the response
                estimated_hours = float(next((line.split(":")[1].strip() for line in response_text.split("\n")
                                              if "estimated hours" in line.lower()), 0))
                skills = [skill.strip() for line in response_text.split("\n")
                          if "skills" in line.lower()
                          for skill in line.split(":")[1].split(",")]
                recommended_priority = next((line.split(":")[1].strip() for line in response_text.split("\n")
                                             if "priority" in line.lower()), priority)
                suggested_deadline = next((line.split(":")[1].strip() for line in response_text.split("\n")
                                           if "deadline" in line.lower()), due_date)
                blockers = [blocker.strip() for line in response_text.split("\n")
                            if "blocker" in line.lower()
                            for blocker in line.split(":")[1].split(",")]
            except Exception as e:
                logger.warning(f"Error parsing LLM response: {str(e)}")
                return task_data

            return {
                **task_data,
                "estimated_hours": estimated_hours,
                "suggested_skills": skills,
                "recommended_priority": recommended_priority,
                "suggested_deadline": suggested_deadline,
                "potential_blockers": blockers
            }
        except Exception as e:
            logger.error(f"Task creation failed: {str(e)}")
            raise

    async def update_task(
        self,
        task_id: str,
        updates: Dict,
        current_state: Dict
    ) -> Dict:
        """Update task with AI validation and recommendations."""
        try:
            # Analyze impact of updates
            impact_analysis = await self.ai_service.generate_response(
                prompt=f"""Analyze the impact of these task updates:
Current State: {current_state}
Proposed Updates: {updates}

Please provide:
1. Impact assessment
2. Suggested adjustments
3. Risk evaluation
4. Timeline implications""",
                context={
                    "system_message": "You are a task management AI assistant. Analyze task updates and assess their impact. Focus on practical implications and provide actionable recommendations."
                }
            )

            validated_updates = {
                **updates,
                "last_updated": datetime.utcnow().isoformat(),
                "update_impact": self._parse_impact_analysis(impact_analysis.get("text", "")),
                "suggested_adjustments": self._parse_adjustments(impact_analysis.get("text", ""))
            }

            return validated_updates
        except Exception as e:
            logger.error(f"Task update failed: {str(e)}")
            raise

    async def plan_task_timeline(
        self,
        task_data: Dict,
        team_capacity: Dict,
        existing_tasks: List[Dict]
    ) -> Dict:
        """Plan optimal task timeline considering team capacity."""
        try:
            timeline = await self.ai_service.generate_response(
                prompt=f"""Plan optimal timeline for this task:
Task: {task_data}
Team Capacity: {team_capacity}
Existing Tasks: {existing_tasks}

Please provide:
1. Suggested start and end dates
2. Key milestones
3. Dependencies schedule
4. Resource allocation plan""",
                context={
                    "system_message": "You are a task management AI assistant specializing in timeline planning. Consider team capacity, dependencies, and existing workload to create realistic schedules."
                }
            )

            return self._parse_timeline_response(timeline.get("text", ""))
        except Exception as e:
            logger.error(f"Task timeline planning failed: {str(e)}")
            raise

    def _parse_impact_analysis(self, response: str) -> Dict:
        """Parse impact analysis from LLM response."""
        try:
            lines = response.split("\n")
            return {
                "impact_level": next((line.split(":")[1].strip() for line in lines
                                      if "impact level" in line.lower()), "unknown"),
                "affected_areas": [area.strip() for line in lines
                                   if "affected area" in line.lower()
                                   for area in line.split(":")[1].split(",")],
                "risks": [risk.strip() for line in lines
                          if "risk" in line.lower()
                          for risk in line.split(":")[1].split(",")]
            }
        except Exception as e:
            logger.warning(f"Error parsing impact analysis: {str(e)}")
            return {}

    def _parse_adjustments(self, response: str) -> List[str]:
        """Parse suggested adjustments from LLM response."""
        try:
            return [adj.strip() for line in response.split("\n")
                    if "adjustment" in line.lower() or "suggestion" in line.lower()
                    for adj in line.split(":")[1].split(",")]
        except Exception as e:
            logger.warning(f"Error parsing adjustments: {str(e)}")
            return []

    def _parse_timeline_response(self, response: str) -> Dict:
        """Parse timeline planning response from LLM."""
        try:
            lines = response.split("\n")
            return {
                "suggested_start_date": next((line.split(":")[1].strip() for line in lines
                                              if "start date" in line.lower()), None),
                "suggested_end_date": next((line.split(":")[1].strip() for line in lines
                                            if "end date" in line.lower()), None),
                "milestones": [milestone.strip() for line in lines
                               if "milestone" in line.lower()
                               for milestone in line.split(":")[1].split(",")],
                "dependencies_schedule": self._parse_dependencies(response),
                "resource_allocation": self._parse_resources(response)
            }
        except Exception as e:
            logger.warning(f"Error parsing timeline response: {str(e)}")
            return {}

    async def delete_task(self, task_id: str, task_data: Dict) -> Dict:
        """Analyze impact of task deletion and provide recommendations."""
        try:
            deletion_impact = await self.ai_service.generate_response(
                prompt="Analyze task deletion impact",
                context={"task": task_data}
            )

            return {
                "can_delete": deletion_impact.get("can_delete", True),
                "impact_assessment": deletion_impact.get("impact_assessment", {}),
                "affected_dependencies": deletion_impact.get("affected_dependencies", []),
                "recommended_actions": deletion_impact.get("recommended_actions", [])
            }
        except Exception as e:
            logger.error(f"Task deletion analysis failed: {str(e)}")
            raise

    def _parse_dependencies(self, response: str) -> Dict[str, List[str]]:
        """Parse dependencies schedule from LLM response."""
        try:
            dependencies = {}
            in_dependencies = False
            for line in response.split("\n"):
                if "dependencies" in line.lower():
                    in_dependencies = True
                    continue
                if in_dependencies and ":" in line:
                    task, deps = line.split(":", 1)
                    dependencies[task.strip()] = [d.strip()
                                                  for d in deps.split(",")]
            return dependencies
        except Exception as e:
            logger.warning(f"Error parsing dependencies: {str(e)}")
            return {}

    def _parse_resources(self, response: str) -> Dict[str, Dict[str, Any]]:
        """Parse resource allocation from LLM response."""
        try:
            resources = {}
            in_resources = False
            current_resource = None
            for line in response.split("\n"):
                if "resource allocation" in line.lower():
                    in_resources = True
                    continue
                if in_resources and line.strip():
                    if ":" in line and not line.startswith(" "):
                        current_resource = line.split(":")[0].strip()
                        resources[current_resource] = {}
                    elif current_resource and ":" in line:
                        key, value = line.strip().split(":", 1)
                        resources[current_resource][key.strip()
                                                    ] = value.strip()
            return resources
        except Exception as e:
            logger.warning(f"Error parsing resources: {str(e)}")
            return {}
