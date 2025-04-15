"""
Agent for generating task reports.
"""
from typing import Dict, Any, Optional, List
import logging
from datetime import datetime, timedelta, timezone

from ai_services.agents.report_agents.base_report_agent import BaseReportAgent
from data_layer.models.report import ReportType
from ai_services.report.data_fetcher import DataFetcherService

logger = logging.getLogger(__name__)


class TaskReportAgent(BaseReportAgent):
    """
    Agent for generating task reports.

    This agent analyzes user task and todo data to generate
    reports with insights and recommendations for task management.
    """

    report_type = ReportType.TASK
    name = "task_report_agent"
    description = "Agent for generating task reports with insights and recommendations"

    def __init__(self):
        """Initialize the task report agent."""
        super().__init__()
        self.data_fetcher = DataFetcherService()

    async def gather_context(
        self,
        user_id: str,
        parameters: Dict[str, Any],
        time_range: Dict[str, str],
        auth_token: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Gather user task data for report generation.

        Parameters:
            user_id (str): User ID to gather data for
            parameters (Dict[str, Any]): Additional parameters for gathering context
            time_range (Dict[str, str]): Time range for data (start_date, end_date)
            auth_token (Optional[str]): Authentication token

        Returns:
            Dict[str, Any]: Context data for report generation
        """
        # Get basic context from parent class
        context = await super().gather_context(user_id, parameters, time_range, auth_token)

        try:
            # Prepare task filters from the incoming parameters
            task_filters = {
                "status": parameters.get("status"),
                "priority": parameters.get("priority"),
                "project_id": parameters.get("project_id")
            }
            # Clean the filters dict to remove any None values
            task_filters = {k: v for k,
                            v in task_filters.items() if v is not None}

            # Fetch task, todo, and project data
            task_data = await self.data_fetcher.fetch_user_data(
                user_id,
                "tasks",
                task_filters,
                time_range,
                auth_token
            )

            todo_data = await self.data_fetcher.fetch_user_data(
                user_id,
                "todos",
                {},
                time_range,
                auth_token
            )

            project_data = await self.data_fetcher.fetch_user_data(
                user_id,
                "projects",
                {},
                time_range,
                auth_token
            )

            # Add data to context
            context["task_data"] = task_data
            context["todo_data"] = todo_data
            context["project_data"] = project_data

            # Extract key metrics from task data
            if task_data and isinstance(task_data.get("tasks"), list):
                tasks = task_data["tasks"]
                completed_tasks = [
                    t for t in tasks if t.get("status") == "Completed"]
                context["tasks_completed"] = len(completed_tasks)
                context["tasks_total"] = len(tasks)
                context["task_completion_rate"] = (
                    len(completed_tasks) / len(tasks) * 100) if tasks else 0
                context["task_overdue"] = len([t for t in tasks if t.get("due_date") and datetime.fromisoformat(
                    t["due_date"].replace('Z', '+00:00')) < datetime.now(timezone.utc) and t.get("status") != "Completed"])

            # Extract key metrics from todo data
            if todo_data and isinstance(todo_data.get("data", {}).get("lists"), list):
                all_todos = []
                for a_list in todo_data["data"]["lists"]:
                    if isinstance(a_list.get("todos"), list):
                        all_todos.extend(a_list["todos"])

                completed_todos = [
                    t for t in all_todos if t.get("is_completed")]
                context["todos_completed"] = len(completed_todos)
                context["todos_total"] = len(all_todos)
                context["todo_completion_rate"] = (
                    len(completed_todos) / len(all_todos) * 100) if all_todos else 0

            # Extract key metrics from project data
            if project_data and isinstance(project_data.get("projects"), list):
                projects = project_data["projects"]
                completed_projects = [
                    p for p in projects if p.get("status") == "completed"]
                context["projects_completed"] = len(completed_projects)
                context["projects_total"] = len(projects)
                context["project_completion_rate"] = (
                    len(completed_projects) / len(projects) * 100) if projects else 0
                context["projects_on_time"] = len([p for p in completed_projects if p.get(
                    "endDate") and p.get("actualEndDate") and p.get("actualEndDate") <= p.get("endDate")])
                context["projects_delayed"] = len([p for p in completed_projects if p.get(
                    "endDate") and p.get("actualEndDate") and p.get("actualEndDate") > p.get("endDate")])

            logger.info(
                f"Successfully gathered and processed task data for user {user_id}")

        except Exception as e:
            logger.error(f"Error gathering task data: {str(e)}")
            context["error"] = str(e)

        return context

    async def _prepare_report_prompt(self, context: Dict[str, Any]) -> str:
        """
        Prepare the prompt for task report generation.

        Parameters:
            context (Dict[str, Any]): Context data for report generation

        Returns:
            str: Formatted prompt for LLM
        """
        time_range = context.get("time_range", {})
        start_date = time_range.get("start_date", "")
        end_date = time_range.get("end_date", "")

        # Extract key metrics for the prompt
        task_completion_rate = context.get("task_completion_rate", "N/A")
        tasks_completed = context.get("tasks_completed", "N/A")
        tasks_total = context.get("tasks_total", "N/A")
        task_overdue = context.get("task_overdue", "N/A")
        avg_task_completion_time = context.get(
            "avg_task_completion_time", "N/A")

        todo_completion_rate = context.get("todo_completion_rate", "N/A")
        todos_completed = context.get("todos_completed", "N/A")
        todos_total = context.get("todos_total", "N/A")

        project_completion_rate = context.get("project_completion_rate", "N/A")
        projects_completed = context.get("projects_completed", "N/A")
        projects_total = context.get("projects_total", "N/A")
        projects_on_time = context.get("projects_on_time", "N/A")
        projects_delayed = context.get("projects_delayed", "N/A")

        # Extract raw data for deeper analysis
        task_data = context.get("task_data", {})
        todo_data = context.get("todo_data", {})
        project_data = context.get("project_data", {})

        prompt = f"""
        Generate a detailed task management report for the user based on their data from {start_date} to {end_date}.
        
        Key metrics:
        - Task Completion Rate: {task_completion_rate}%
        - Tasks Completed: {tasks_completed} out of {tasks_total}
        - Overdue Tasks: {task_overdue}
        - Average Task Completion Time: {avg_task_completion_time} hours
        
        - Todo Completion Rate: {todo_completion_rate}%
        - Todos Completed: {todos_completed} out of {todos_total}
        
        - Project Completion Rate: {project_completion_rate}%
        - Projects Completed: {projects_completed} out of {projects_total}
        - Projects On Time: {projects_on_time}
        - Projects Delayed: {projects_delayed}
        
        The report should include the following sections:
        1. Task Management Summary - A high-level overview of the user's task management during this period
        2. Task Analysis - Analysis of task completion patterns, efficiency, and bottlenecks
        3. Todo Analysis - Analysis of todo completion and organization
        4. Project Analysis - Analysis of project progress and timeliness
        5. Recommendations - Actionable suggestions for improving task management
        
        Raw data:
        Task Data: {task_data}
        Todo Data: {todo_data}
        Project Data: {project_data}
        
        Return the report as a JSON with the following structure:
        {{
            "summary": "Brief summary of key findings",
            "content": {{
                "task_efficiency_score": 85,  // Overall task efficiency score out of 100
                "key_metrics": {{
                    // Key metrics extracted from the data
                }},
                "insights": [
                    // List of key insights
                ],
                "bottlenecks": [
                    // List of identified bottlenecks
                ],
                "recommendations": [
                    // List of recommendations
                ]
            }},
            "sections": [
                {{
                    "title": "Task Management Summary",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Task Analysis",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Todo Analysis",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Project Analysis",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Recommendations",
                    "content": "Detailed recommendations...",
                    "type": "text"
                }}
            ]
        }}
        """

        return prompt
