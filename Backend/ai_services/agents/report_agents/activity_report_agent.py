"""
Agent for generating activity reports.
"""
from typing import Dict, Any, Optional, List
import logging
from datetime import datetime, timedelta, timezone

from ai_services.agents.report_agents.base_report_agent import BaseReportAgent
from data_layer.models.report import ReportType
from ai_services.report.data_fetcher import DataFetcherService
from core.mcp_state import get_mcp_client

logger = logging.getLogger(__name__)


class ActivityReportAgent(BaseReportAgent):
    """
    Agent for generating activity reports.

    This agent analyzes user activity data and generates
    reports with insights and recommendations.
    """

    report_type = ReportType.ACTIVITY
    name = "activity_report_agent"
    description = "Agent for generating activity reports"

    def __init__(self):
        """Initialize the activity report agent."""
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
        Gather user activity data for report generation.

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
            # Fetch data using the DataFetcherService
            metric_types = ["activity", "tasks", "calendar"]
            metrics = await self.data_fetcher.fetch_metrics(
                user_id,
                metric_types,
                time_range,
                auth_token
            )

            # Add raw data to context for the prompt
            context.update({
                "activity_data": metrics.get("activity", {}),
                "task_data": metrics.get("tasks", {}),
                "calendar_data": metrics.get("calendar", {}),
            })

            # Process raw data to extract key metrics for the prompt
            if isinstance(metrics.get("tasks", {}).get("tasks"), list):
                tasks = metrics["tasks"]["tasks"]
                completed_tasks = [
                    t for t in tasks if t.get("status") == "Completed"]
                context["tasks_completed"] = len(completed_tasks)
                context["tasks_total"] = len(tasks)
                context["task_overdue"] = len([
                    t for t in tasks if t.get("due_date") and
                    datetime.fromisoformat(t["due_date"].replace(
                        'Z', '+00:00')) < datetime.now(timezone.utc)
                    and t.get("status") != "Completed"
                ])

            if isinstance(metrics.get("calendar", {}).get("events"), list):
                events = metrics["calendar"]["events"]
                context["meeting_count"] = len(
                    [e for e in events if e.get("type") == "Meeting"])
                context["meeting_time"] = sum(
                    (datetime.fromisoformat(
                        e["end_time"]) - datetime.fromisoformat(e["start_time"])).total_seconds() / 60
                    for e in events if e.get("type") == "Meeting" and "start_time" in e and "end_time" in e
                )

            logger.info(
                f"Successfully gathered activity data for user {user_id}")

        except Exception as e:
            logger.error(f"Error gathering activity data: {str(e)}")
            context["error"] = str(e)

        return context

    async def _prepare_report_prompt(self, context: Dict[str, Any]) -> str:
        """
        Prepare the prompt for activity report generation.

        Parameters:
            context (Dict[str, Any]): Context data for report generation

        Returns:
            str: Formatted prompt for LLM
        """
        time_range = context.get("time_range", {})
        start_date = time_range.get("start_date", "")
        end_date = time_range.get("end_date", "")

        # Key metrics calculated in gather_context
        tasks_completed = context.get("tasks_completed", 0)
        tasks_total = context.get("tasks_total", 0)
        tasks_overdue = context.get("task_overdue", 0)
        meeting_count = context.get("meeting_count", 0)
        meeting_time = context.get("meeting_time", 0)

        # Raw data for deeper analysis by the LLM
        activity_data = context.get("activity_data", {})
        task_data = context.get("task_data", {})
        calendar_data = context.get("calendar_data", {})

        prompt = f"""
        Generate a detailed activity report for the user based on their data from {start_date} to {end_date}.
        
        Key Metrics:
        - Tasks Completed: {tasks_completed} out of {tasks_total}
        - Overdue Tasks: {tasks_overdue}
        - Meetings Attended: {meeting_count}
        - Total Meeting Time: {meeting_time:.0f} minutes

        The report should analyze the user's activity patterns, task completion, and calendar usage.
        
        Include the following sections:
        1. Activity Summary - A high-level overview of the user's activity during this period
        2. Task Analysis - Insights into task completion patterns, efficiency, and areas for improvement
        3. Calendar Analysis - Analysis of time management, meeting patterns, and scheduling efficiency
        4. Recommendations - Actionable suggestions for improving productivity and time management
        
        Raw Activity Data for Analysis:
        {activity_data}
        
        Raw Task Data for Analysis:
        {task_data}
        
        Raw Calendar Data for Analysis:
        {calendar_data}
        
        Return the report as a JSON with the following structure:
        {{
            "summary": "Brief summary of key findings",
            "content": {{
                "activity_score": 85,  // Overall activity score out of 100
                "key_metrics": {{
                    "tasks_completed": "{tasks_completed} out of {tasks_total}",
                    "overdue_tasks": {tasks_overdue},
                    "meetings_attended": {meeting_count},
                    "total_meeting_time_minutes": {meeting_time:.0f}
                }},
                "insights": [
                    // List of key insights
                ]
            }},
            "sections": [
                {{
                    "title": "Activity Summary",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Task Analysis",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Calendar Analysis",
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
