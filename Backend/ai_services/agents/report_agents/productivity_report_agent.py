"""
Agent for generating productivity reports.
"""
from typing import Dict, Any, Optional, List
import logging
from datetime import datetime, timedelta, timezone

from ai_services.agents.report_agents.base_report_agent import BaseReportAgent
from data_layer.models.report import ReportType
from ai_services.report.data_fetcher import DataFetcherService

logger = logging.getLogger(__name__)


class ProductivityReportAgent(BaseReportAgent):
    """
    Agent for generating productivity reports.

    This agent analyzes user productivity data and generates
    reports with insights and recommendations for improvement.
    """

    report_type = ReportType.PRODUCTIVITY
    name = "productivity_report_agent"
    description = "Agent for generating productivity reports with insights and recommendations"

    def __init__(self):
        """Initialize the productivity report agent."""
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
        Gather user productivity data for report generation.

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
            # Fetch multiple data types in parallel
            metric_types = [
                "productivity",
                "focus",
                "tasks",
                "todos",
                "calendar",
                "dashboard"
            ]

            metrics = await self.data_fetcher.fetch_metrics(
                user_id,
                metric_types,
                time_range,
                auth_token
            )

            # Add metrics to context
            context.update(metrics)

            # --- Process Raw Data to Calculate Metrics ---

            # Calculate average productivity score from raw data
            if "productivity" in metrics and "daily_scores" in metrics.get("productivity", {}):
                daily_scores = metrics["productivity"]["daily_scores"]
                if daily_scores:
                    avg_score = sum(
                        score for _, score in daily_scores.items()) / len(daily_scores)
                    context["avg_productivity_score"] = round(avg_score, 2)

            # Extract focus time metrics
            if "focus" in metrics and "total_focus_time" in metrics.get("focus", {}):
                focus_data = metrics["focus"]
                context["total_focus_time"] = focus_data.get(
                    "total_focus_time", 0)
                daily_focus = focus_data.get("daily_focus_time", {})
                if daily_focus:
                    avg_focus = sum(
                        time for _, time in daily_focus.items()) / len(daily_focus)
                    context["avg_daily_focus_time"] = round(avg_focus / 60, 2)

            # Process raw task list to calculate metrics
            if "tasks" in metrics and isinstance(metrics.get("tasks", {}).get("tasks"), list):
                tasks = metrics["tasks"]["tasks"]
                completed_tasks = [
                    t for t in tasks if t.get("status") == "Completed"]
                context["tasks_completed"] = len(completed_tasks)
                context["tasks_total"] = len(tasks)
                context["task_completion_rate"] = (
                    len(completed_tasks) / len(tasks) * 100) if tasks else 0

            # Process raw todo lists to calculate metrics
            if "todos" in metrics and isinstance(metrics.get("todos", {}).get("data", {}).get("lists"), list):
                all_todos = []
                for a_list in metrics["todos"]["data"]["lists"]:
                    if isinstance(a_list.get("todos"), list):
                        all_todos.extend(a_list["todos"])

                completed_todos = [
                    t for t in all_todos if t.get("is_completed")]
                context["todos_completed_count"] = len(completed_todos)
                context["todos_total_count"] = len(all_todos)

            # Extract calendar metrics
            if "calendar" in metrics and "events" in metrics.get("calendar", {}):
                calendar_data = metrics["calendar"]
                meeting_events = [e for e in calendar_data.get(
                    "events", []) if e.get("type") == "Meeting"]
                context["meeting_time"] = sum(
                    (datetime.fromisoformat(
                        e["end_time"]) - datetime.fromisoformat(e["start_time"])).total_seconds() / 60
                    for e in meeting_events if "start_time" in e and "end_time" in e
                )
                context["meeting_count"] = len(meeting_events)

            logger.info(
                f"Successfully gathered and processed productivity data for user {user_id}")

        except Exception as e:
            logger.error(f"Error gathering productivity data: {str(e)}")
            context["error"] = str(e)

        return context

    async def _prepare_report_prompt(self, context: Dict[str, Any]) -> str:
        """
        Prepare the prompt for productivity report generation.

        Parameters:
            context (Dict[str, Any]): Context data for report generation

        Returns:
            str: Formatted prompt for LLM
        """
        time_range = context.get("time_range", {})
        start_date = time_range.get("start_date", "")
        end_date = time_range.get("end_date", "")

        # Extract key metrics for the prompt
        avg_productivity_score = context.get("avg_productivity_score", "N/A")
        avg_daily_focus_time = context.get("avg_daily_focus_time", "N/A")
        task_completion_rate = context.get("task_completion_rate", 0.0)
        tasks_completed = context.get("tasks_completed", 0)
        tasks_total = context.get("tasks_total", 0)
        meeting_time = context.get("meeting_time", 0)
        meeting_count = context.get("meeting_count", 0)

        # Raw data for deeper analysis
        productivity_data = context.get("productivity", {})
        focus_data = context.get("focus", {})
        task_data = context.get("tasks", {})
        todo_data = context.get("todos", {})
        calendar_data = context.get("calendar", {})
        dashboard_data = context.get("dashboard", {})

        prompt = f"""
        Generate a detailed productivity report for the user based on their data from {start_date} to {end_date}.
        
        Key metrics:
        - Average Productivity Score: {avg_productivity_score}
        - Average Daily Focus Time: {avg_daily_focus_time} hours
        - Task Completion Rate: {task_completion_rate:.2f}%
        - Tasks Completed: {tasks_completed} out of {tasks_total}
        - Meeting Time: {meeting_time:.0f} minutes across {meeting_count} meetings
        
        The report should include the following sections:
        1. Productivity Summary - A high-level overview of the user's productivity during this period
        2. Focus Time Analysis - Analysis of focus time patterns, effectiveness, and areas for improvement
        3. Task Completion Analysis - Insights into task completion patterns, efficiency, and bottlenecks
        4. Time Management - Analysis of calendar usage, meeting patterns, and time allocation
        5. Recommendations - Actionable suggestions for improving productivity based on the data
        
        Raw data:
        Productivity Data: {productivity_data}
        Focus Data: {focus_data}
        Task Data: {task_data}
        Todo Data: {todo_data}
        Calendar Data: {calendar_data}
        Dashboard Data: {dashboard_data}
        
        Return the report as a JSON with the following structure:
        {{
            "summary": "Brief summary of key findings",
            "content": {{
                "productivity_score": 85,  // Overall productivity score out of 100
                "key_metrics": {{
                    "average_productivity_score": "N/A",
                    "average_daily_focus_time_hours": "N/A",
                    "task_completion_rate": "{task_completion_rate:.2f}%",
                    "tasks_completed": "{tasks_completed} out of {tasks_total}",
                    "meeting_time_minutes": {meeting_time:.0f},
                    "number_of_meetings": {meeting_count}
                }},
                "insights": [
                    // List of key insights
                ],
                "areas_for_improvement": [
                    // List of areas that need improvement
                ]
            }},
            "sections": [
                {{
                    "title": "Productivity Summary",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Focus Time Analysis",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Task Completion Analysis",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Time Management",
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
