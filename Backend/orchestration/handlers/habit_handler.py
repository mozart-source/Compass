from typing import Dict, List, Any, Optional
import logging
from Backend.orchestration.handlers.base_handler import BaseHandler
from Backend.data_layer.repositories.daily_habits_repository import DailyHabitRepository

logger = logging.getLogger(__name__)

class HabitHandler(BaseHandler):
    async def enrich_context(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """
        Add active habits and streaks to the context.
        """
        habit_repo = DailyHabitRepository(self.db)
        active_habits = await habit_repo.get_active_habits(
            user_id=context["user_id"]
        )
        context["active_habits"] = [
            {
                "habit_name": habit.habit_name,
                "description": habit.description,
                "current_streak": habit.current_streak,
                "longest_streak": habit.longest_streak,
                "is_completed": habit.is_completed,
                "last_completed_date": habit.last_completed_date.isoformat() if habit.last_completed_date else None,
                "streak_start_date": habit.streak_start_date.isoformat() if habit.streak_start_date else None
            } for habit in active_habits
        ]
        return context
