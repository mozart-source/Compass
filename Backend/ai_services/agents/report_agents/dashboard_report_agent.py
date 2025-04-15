"""
Agent for generating dashboard reports.
"""
from typing import Dict, Any, Optional, List
import logging
from datetime import datetime, timedelta, timezone

from ai_services.agents.report_agents.base_report_agent import BaseReportAgent
from data_layer.models.report import ReportType
from ai_services.report.data_fetcher import DataFetcherService
from core.mcp_state import get_mcp_client

logger = logging.getLogger(__name__)


class DashboardReportAgent(BaseReportAgent):
    """
    Agent for generating dashboard reports.

    This agent creates comprehensive dashboard reports that provide
    an overview of user activity across multiple aspects of the platform,
    similar to a summary report but with a focus on dashboard metrics.
    """

    report_type = ReportType.DASHBOARD
    name = "dashboard_report_agent"
    description = "Agent for generating comprehensive dashboard reports across multiple data types"

    def __init__(self):
        """Initialize the dashboard report agent."""
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
        Gather comprehensive user data for dashboard report generation.

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
            # Fetch data from all available sources
            metric_types = [
                "activity",
                "productivity",
                "focus",
                "tasks",
                "todos",
                "habits",
                "calendar",
                "dashboard"
            ]

            metrics = await self.data_fetcher.fetch_metrics(
                user_id,
                metric_types,
                time_range,
                auth_token
            )

            context.update(metrics)

            # --- Process Raw Data to Calculate Key Metrics ---
            self._calculate_and_add_metrics(context, metrics)

            logger.info(
                f"Successfully gathered and processed dashboard data for user {user_id}")

        except Exception as e:
            logger.error(f"Error gathering dashboard data: {str(e)}")
            context["error"] = str(e)

        return context

    def _calculate_and_add_metrics(self, context: Dict[str, Any], metrics: Dict[str, Any]) -> None:
        """Process raw data from various sources to calculate metrics for the context."""

        # Process task data
        if "tasks" in metrics and isinstance(metrics.get("tasks", {}).get("tasks"), list):
            tasks = metrics["tasks"]["tasks"]
            completed_tasks = [
                t for t in tasks if t.get("status") == "Completed"]
            context["tasks_completed"] = len(completed_tasks)
            context["tasks_total"] = len(tasks)
            context["task_completion_rate"] = (
                len(completed_tasks) / len(tasks) * 100) if tasks else 0

        # Process habit data
        if "habits" in metrics and isinstance(metrics.get("habits", {}).get("data", {}).get("habits"), list):
            habits = metrics["habits"]["data"]["habits"]
            total_habits = len(habits)
            if total_habits > 0:
                completed_habits = len(
                    [h for h in habits if h.get("is_completed")])
                context["habit_completion_rate"] = (
                    completed_habits / total_habits) * 100
            else:
                context["habit_completion_rate"] = 0

        # Process focus data
        if "focus" in metrics and metrics.get("focus"):
            context["total_focus_time"] = metrics["focus"].get(
                "total_focus_time", 0)

        # Process calendar data
        if "calendar" in metrics and isinstance(metrics.get("calendar", {}).get("events"), list):
            events = metrics["calendar"]["events"]
            meeting_events = [e for e in events if e.get("type") == "Meeting"]
            context["meeting_count"] = len(meeting_events)

    async def _prepare_report_prompt(self, context: Dict[str, Any]) -> str:
        """
        Prepare the prompt for dashboard report generation.

        Parameters:
            context (Dict[str, Any]): Context data for report generation

        Returns:
            str: Formatted prompt for LLM
        """
        time_range = context.get("time_range", {})
        start_date = time_range.get("start_date", "")
        end_date = time_range.get("end_date", "")

        # Extract calculated metrics for a clean prompt
        task_completion_rate = context.get("task_completion_rate", 0.0)
        tasks_completed = context.get("tasks_completed", 0)
        tasks_total = context.get("tasks_total", 0)
        habit_completion_rate = context.get("habit_completion_rate", 0.0)
        total_focus_time = context.get(
            "total_focus_time", 0) / 3600  # Convert to hours
        meeting_count = context.get("meeting_count", 0)

        # Extract raw data for deeper analysis
        activity_data = context.get("activity", {})
        productivity_data = context.get("productivity", {})
        habit_data = context.get("habits", {})
        task_data = context.get("tasks", {})

        prompt = f"""
        Generate a comprehensive dashboard report for the user based on their data from {start_date} to {end_date}.
        The report should be a narrative summary of the key metrics and insights available on their dashboard.
        
        Key Metrics:
        - Task Completion: {tasks_completed} of {tasks_total} tasks completed ({task_completion_rate:.2f}%)
        - Habit Completion Rate: {habit_completion_rate:.2f}%
        - Total Focus Time: {total_focus_time:.2f} hours
        - Meetings Attended: {meeting_count}
        
        The report should include the following sections:
        1. Executive Summary - A high-level overview of the user's activity and achievements.
        2. Productivity Overview - Summary of productivity scores, focus time, and task completion.
        3. Habit Consistency - Analysis of habit formation and streaks.
        4. Time Management - Insights from calendar events and meeting patterns.
        5. Key Recommendations - Actionable suggestions based on the dashboard data.
        
        Use the following raw data only for deeper analysis if needed, but primarily rely on the key metrics provided above:
        Activity Data: {activity_data}
        Productivity Data: {productivity_data}
        Habit Data: {habit_data}
        Task Data: {task_data}
        
        Return the report as a JSON with the following structure:
        {{
            "summary": "Brief executive summary of key findings from the dashboard.",
            "content": {{
                "overall_score": 85,  // An overall user performance score out of 100 based on dashboard data
                "key_insights": [
                    "Insight about productivity...",
                    "Insight about habits...",
                    "Insight about time management..."
                ],
                "recommendations": [
                    "Actionable recommendation based on insights..."
                ]
            }},
            "sections": [
                {{
                    "title": "Executive Summary",
                    "content": "Detailed summary...",
                    "type": "text"
                }},
                {{
                    "title": "Productivity Overview",
                    "content": "Detailed overview...",
                    "type": "text"
                }},
                {{
                    "title": "Habit Consistency",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Time Management",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Key Recommendations",
                    "content": "Detailed recommendations...",
                    "type": "text"
                }}
            ]
        }}
        """

        return prompt
