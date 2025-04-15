"""
Agent for generating summary reports.
"""
from typing import Dict, Any, Optional, List
import logging
from datetime import datetime, timedelta

from ai_services.agents.report_agents.base_report_agent import BaseReportAgent
from data_layer.models.report import ReportType
from ai_services.report.data_fetcher import DataFetcherService

logger = logging.getLogger(__name__)


class SummaryReportAgent(BaseReportAgent):
    """
    Agent for generating summary reports.

    This agent creates comprehensive summary reports that provide
    an overview of user activity across multiple aspects of the platform.
    """

    report_type = ReportType.SUMMARY
    name = "summary_report_agent"
    description = "Agent for generating comprehensive summary reports across multiple data types"

    def __init__(self):
        """Initialize the summary report agent."""
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
        Gather comprehensive user data for summary report generation.

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

            # Also fetch dashboard data which may contain additional metrics
            dashboard_data = await self.data_fetcher.fetch_dashboard_data(
                user_id,
                auth_token
            )

            if dashboard_data:
                metrics["dashboard"] = dashboard_data

            # Add all metrics to context
            context.update(metrics)

            # Extract key metrics for easier access in prompt generation
            self._extract_key_metrics(context, metrics)

            logger.info(
                f"Successfully gathered summary data for user {user_id}")

        except Exception as e:
            logger.error(f"Error gathering summary data: {str(e)}")
            context["error"] = str(e)

        return context

    def _extract_key_metrics(self, context: Dict[str, Any], metrics: Dict[str, Any]) -> None:
        """Process raw data to calculate and extract key metrics for the summary report."""

        # --- Activity ---
        if "activity" in metrics and metrics.get("activity"):
            activity = metrics["activity"]
            context["active_days"] = activity.get("active_days", 0)
            context["active_hours"] = activity.get("active_hours", 0)

        # --- Productivity ---
        if "productivity" in metrics and "daily_scores" in metrics.get("productivity", {}):
            daily_scores = metrics["productivity"]["daily_scores"]
            if daily_scores:
                avg_score = sum(
                    score for _, score in daily_scores.items()) / len(daily_scores)
                context["avg_productivity_score"] = round(avg_score, 2)

        # --- Focus ---
        if "focus" in metrics and metrics.get("focus"):
            focus = metrics["focus"]
            context["total_focus_time"] = focus.get("total_focus_time", 0)
            context["focus_sessions"] = focus.get("sessions", 0)

        # --- Tasks ---
        if "tasks" in metrics and isinstance(metrics.get("tasks", {}).get("tasks"), list):
            tasks = metrics["tasks"]["tasks"]
            completed_tasks = [
                t for t in tasks if t.get("status") == "Completed"]
            context["tasks_completed"] = len(completed_tasks)
            context["tasks_total"] = len(tasks)
            context["task_completion_rate"] = (
                len(completed_tasks) / len(tasks) * 100) if tasks else 0

        # --- Habits ---
        if "habits" in metrics and isinstance(metrics.get("habits", {}).get("data", {}).get("habits"), list):
            habits = metrics["habits"]["data"]["habits"]
            total_habits = len(habits)
            if total_habits > 0:
                completed_habits = len(
                    [h for h in habits if h.get("is_completed")])
                context["habit_completion_rate"] = (
                    completed_habits / total_habits) * 100
                context["habits_completed"] = completed_habits
                context["habits_total"] = total_habits
            else:
                context["habit_completion_rate"] = 0
                context["habits_completed"] = 0
                context["habits_total"] = 0

        # --- Todos ---
        if "todos" in metrics and isinstance(metrics.get("todos", {}).get("data", {}).get("lists"), list):
            all_todos = []
            for a_list in metrics["todos"]["data"]["lists"]:
                if isinstance(a_list.get("todos"), list):
                    all_todos.extend(a_list["todos"])

            completed_todos = [t for t in all_todos if t.get("is_completed")]
            context["todos_completed_count"] = len(completed_todos)
            context["todos_total_count"] = len(all_todos)

        # --- Calendar ---
        if "calendar" in metrics and isinstance(metrics.get("calendar", {}).get("events"), list):
            calendar_events = metrics["calendar"]["events"]
            meeting_events = [
                e for e in calendar_events if e.get("type") == "Meeting"]
            context["meeting_time"] = sum(
                (datetime.fromisoformat(
                    e["end_time"]) - datetime.fromisoformat(e["start_time"])).total_seconds() / 60
                for e in meeting_events if "start_time" in e and "end_time" in e
            )
            context["meeting_count"] = len(meeting_events)

    async def _prepare_report_prompt(self, context: Dict[str, Any]) -> str:
        """
        Prepare the prompt for summary report generation.

        Parameters:
            context (Dict[str, Any]): Context data for report generation

        Returns:
            str: Formatted prompt for LLM
        """
        time_range = context.get("time_range", {})
        start_date = time_range.get("start_date", "")
        end_date = time_range.get("end_date", "")

        # Extract key metrics for the prompt
        active_days = context.get("active_days", "N/A")
        activity_trend = context.get("activity_trend", "N/A")
        avg_productivity_score = context.get("avg_productivity_score", "N/A")
        avg_daily_focus_time = context.get("avg_daily_focus_time", "N/A")
        task_completion_rate = context.get("task_completion_rate", 0)
        tasks_completed = context.get("tasks_completed", 0)
        habit_completion_rate = context.get("habit_completion_rate", 0)
        meeting_time = context.get("meeting_time", 0)
        meeting_count = context.get("meeting_count", 0)
        project_completion_rate = context.get("project_completion_rate", "N/A")
        workflows_executed = context.get("workflows_executed", "N/A")

        # Extract raw data for deeper analysis
        activity_data = context.get("activity", {})
        productivity_data = context.get("productivity", {})
        focus_data = context.get("focus", {})
        task_data = context.get("tasks", {})
        todo_data = context.get("todos", {})
        habit_data = context.get("habits", {})
        calendar_data = context.get("calendar", {})
        project_data = context.get("projects", {})
        workflow_data = context.get("workflow", {})
        dashboard_data = context.get("dashboard", {})

        prompt = f"""
        Generate a comprehensive summary report for the user based on their data from {start_date} to {end_date}.
        
        Key metrics across domains:
        - Active Days: {active_days} days
        - Activity Trend: {activity_trend}
        - Average Productivity Score: {avg_productivity_score}
        - Average Daily Focus Time: {avg_daily_focus_time} hours
        - Task Completion Rate: {task_completion_rate:.2f}%
        - Tasks Completed: {tasks_completed}
        - Habit Completion Rate: {habit_completion_rate:.2f}%
        - Meeting Time: {meeting_time:.0f} minutes across {meeting_count} meetings
        - Project Completion Rate: {project_completion_rate}%
        - Workflows Executed: {workflows_executed}
        
        The report should include the following sections:
        1. Executive Summary - A high-level overview of the user's activity and achievements
        2. Activity Analysis - Analysis of user activity patterns and trends
        3. Productivity Overview - Summary of productivity metrics and focus time
        4. Task & Project Management - Overview of task and project completion
        5. Habit Building - Summary of habit formation and consistency
        6. Time Management - Analysis of calendar usage and time allocation
        7. Key Achievements - Highlight of major accomplishments during this period
        8. Areas for Improvement - Identification of areas that need attention
        9. Recommendations - Actionable suggestions across different domains
        
        Raw data:
        Activity Data: {activity_data}
        Productivity Data: {productivity_data}
        Focus Data: {focus_data}
        Task Data: {task_data}
        Todo Data: {todo_data}
        Habit Data: {habit_data}
        Calendar Data: {calendar_data}
        Project Data: {project_data}
        Workflow Data: {workflow_data}
        Dashboard Data: {dashboard_data}
        
        Return the report as a JSON with the following structure:
        {{
            "summary": "Brief executive summary of key findings",
            "content": {{
                "overall_score": 85,  // Overall user performance score out of 100
                "key_metrics": {{
                    // Key metrics extracted from the data across domains
                }},
                "achievements": [
                    // List of key achievements
                ],
                "areas_for_improvement": [
                    // List of areas that need improvement
                ],
                "recommendations": [
                    // List of recommendations across domains
                ]
            }},
            "sections": [
                {{
                    "title": "Executive Summary",
                    "content": "Detailed summary...",
                    "type": "text"
                }},
                {{
                    "title": "Activity Analysis",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Productivity Overview",
                    "content": "Detailed overview...",
                    "type": "text"
                }},
                {{
                    "title": "Task & Project Management",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Habit Building",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Time Management",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Key Achievements",
                    "content": "Detailed achievements...",
                    "type": "text"
                }},
                {{
                    "title": "Areas for Improvement",
                    "content": "Detailed areas...",
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
