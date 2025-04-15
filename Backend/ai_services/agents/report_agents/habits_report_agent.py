"""
Agent for generating habits reports.
"""
from typing import Dict, Any, Optional, List
import logging
from datetime import datetime, timedelta, timezone

from ai_services.agents.report_agents.base_report_agent import BaseReportAgent
from data_layer.models.report import ReportType
from ai_services.report.data_fetcher import DataFetcherService

logger = logging.getLogger(__name__)


class HabitsReportAgent(BaseReportAgent):
    """
    Agent for generating habits reports.

    This agent analyzes user habits data and generates
    reports with insights and recommendations for habit building.
    """

    report_type = ReportType.HABITS
    name = "habits_report_agent"
    description = "Agent for generating habits reports with insights and recommendations"

    def __init__(self):
        """Initialize the habits report agent."""
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
        Gather and process user habits data for report generation.
        """
        context = await super().gather_context(user_id, parameters, time_range, auth_token)

        try:
            habits_data = await self.data_fetcher.fetch_user_data(
                user_id,
                "habits",
                parameters,  # Pass parameters for potential future filtering
                time_range,
                auth_token,
            )

            context["habits_data"] = habits_data

            # Extract habits list safely from nested structure
            habits_list = habits_data.get("data", {}).get("habits", [])

            # Check if habits data is valid
            if not habits_list or not isinstance(habits_list, list):
                logger.warning(
                    f"Habits data for user {user_id} is missing or not in expected format.")
                return context

            # Calculate metrics
            total_habits = len(habits_list)
            if total_habits == 0:
                context["habit_completion_rate"] = 0
                context["habit_streaks"] = {
                    "max_streak": 0, "average_streak": 0}
                context["habit_categories"] = {}
                context["habit_consistency"] = {}
                context["top_habits"] = []
                context["bottom_habits"] = []
            else:
                # 1. Calculate Overall Completion Rate
                total_completion = sum(
                    h.get("completion_rate", 0) for h in habits_list)
                context["habit_completion_rate"] = total_completion / \
                    total_habits

                # 2. Analyze Streaks
                all_streaks = [h.get("streak", 0) for h in habits_list]
                context["habit_streaks"] = {
                    "max_streak": max(all_streaks) if all_streaks else 0,
                    "average_streak": sum(all_streaks) / len(all_streaks) if all_streaks else 0
                }

                # 3. Analyze Categories
                categories = {}
                for habit in habits_list:
                    cat = habit.get("category", "Uncategorized")
                    if cat not in categories:
                        categories[cat] = {"completions": [], "count": 0}
                    categories[cat]["completions"].append(
                        habit.get("completion_rate", 0))
                    categories[cat]["count"] += 1

                category_analysis = {}
                for cat, data in categories.items():
                    avg_completion = sum(data["completions"]) / \
                        data["count"] if data["count"] > 0 else 0
                    category_analysis[cat] = {
                        "average_completion_rate": round(avg_completion, 2),
                        "habit_count": data["count"]
                    }
                context["habit_categories"] = category_analysis

                # Placeholder for consistency analysis
                context["habit_consistency"] = {}

                # 4. Get Top and Bottom Habits
                if habits_list:
                    sorted_habits = sorted(
                        habits_list,
                        key=lambda h: h.get("completion_rate", 0),
                        reverse=True
                    )
                    context["top_habits"] = sorted_habits[:5]
                    context["bottom_habits"] = sorted_habits[-5:]

                logger.info(
                    f"Successfully gathered and processed habits data for user {user_id}")

        except Exception as e:
            logger.error(f"Error gathering habits data: {str(e)}")
            context["error"] = str(e)

        return context

    async def _prepare_report_prompt(self, context: Dict[str, Any]) -> str:
        """
        Prepare the prompt for habits report generation.

        Parameters:
            context (Dict[str, Any]): Context data for report generation

        Returns:
            str: Formatted prompt for LLM
        """
        time_range = context.get("time_range", {})
        start_date = time_range.get("start_date", "")
        end_date = time_range.get("end_date", "")

        # Extract key metrics for the prompt, using numeric defaults for formatted values
        overall_completion_rate = context.get("habit_completion_rate", 0.0)
        total_habits = context.get("total_habits", 0)
        completed_habits = context.get("completed_habits", 0)
        longest_streak = context.get("habit_streaks", {}).get("max_streak", 0)
        average_streak = context.get(
            "habit_streaks", {}).get("average_streak", 0.0)
        habit_completion_by_category = context.get(
            "habit_completion_by_category", {})

        # Format top habits for the prompt
        top_habits_str = "\n".join([
            f"- {h.get('name', 'Unknown')}: {h.get('completion_rate', 0)}% completion rate, {h.get('streak', 0)}-day streak"
            for h in context.get("top_habits", [])
        ]) if context.get("top_habits", []) else "No habit data available"

        # Format bottom habits for the prompt
        bottom_habits_str = "\n".join([
            f"- {h.get('name', 'Unknown')}: {h.get('completion_rate', 0)}% completion rate"
            for h in context.get("bottom_habits", [])
        ]) if context.get("bottom_habits", []) else "No habit data available"

        prompt = f"""
        Generate a detailed habits report for the user based on their data from {start_date} to {end_date}.
        
        Key metrics:
        - Overall Habit Completion Rate: {overall_completion_rate:.2f}%
        - Total Habits: {total_habits}
        - Completed Habits: {completed_habits}
        - Longest Streak Achieved: {longest_streak} days
        - Average Streak Length: {average_streak:.2f} days
        
        Top performing habits:
        {top_habits_str}
        
        Habits needing improvement:
        {bottom_habits_str}
        
        The report should include the following sections:
        1. Habits Summary - A high-level overview of the user's habit performance during this period
        2. Habit Streaks Analysis - Analysis of habit streaks, consistency, and patterns
        3. Category Analysis - Insights into performance across different habit categories
        4. Time of Day Analysis - When the user is most successful at completing habits
        5. Recommendations - Actionable suggestions for improving habit consistency and building new habits
        
        Raw data for context:
        {context.get("habits_data", {})}
        
        Return the report as a JSON with the following structure:
        {{
            "summary": "Brief summary of key findings",
            "content": {{
                "overall_score": {overall_completion_rate},  // Overall habits score out of 100
                "key_metrics": {{
                    "overall_completion_rate": {overall_completion_rate},
                    "total_habits": {total_habits},
                    "completed_habits": {completed_habits},
                    "longest_streak": {longest_streak},
                    "average_streak": {average_streak}
                }},
                "insights": [
                    // List of key insights
                ],
                "habit_recommendations": [
                    // List of habit-specific recommendations
                ]
            }},
            "sections": [
                {{
                    "title": "Habits Summary",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Habit Streaks Analysis",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Category Analysis",
                    "content": "Detailed analysis...",
                    "type": "text"
                }},
                {{
                    "title": "Time of Day Analysis",
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
