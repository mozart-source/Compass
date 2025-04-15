from data_layer.repos.base_repo import BaseMongoRepository
from data_layer.models.focus_model import FocusSession, FocusSettings
from datetime import datetime, timedelta, timezone
from pymongo import ReturnDocument
from bson.objectid import ObjectId
import logging

logger = logging.getLogger(__name__)


class FocusSessionRepository(BaseMongoRepository[FocusSession]):
    def __init__(self):
        super().__init__(FocusSession)

    def find_by_user(self, user_id: str, limit: int = 100):
        return self.find_many({"user_id": user_id}, limit=limit, sort=[("start_time", -1)])

    def find_active_session(self, user_id: str):
        return self.find_one({"user_id": user_id, "status": "active"})

    def get_stats(self, user_id: str, days: int = 30):
        since = datetime.utcnow() - timedelta(days=days)
        sessions = self.find_many(
            {"user_id": user_id, "start_time": {"$gte": since}})
        total = sum((s.duration or 0)
                    for s in sessions if s.status == "completed")
        current_streak, longest_streak = self._calculate_streaks(sessions)
        return {"total_focus_seconds": total, "streak": current_streak, "longest_streak": longest_streak, "sessions": len(sessions)}

    def _calculate_streaks(self, sessions):
        """
        Returns (current_streak, longest_streak):
        - current_streak: consecutive days up to today with at least one completed session
        - longest_streak: max consecutive days with at least one completed session
        """
        # Filter only completed sessions and get their dates (in UTC, date only)
        completed_dates = set()
        for s in sessions:
            if s.status == "completed" and s.start_time:
                dt = s.start_time
                if dt.tzinfo is None:
                    dt = dt.replace(tzinfo=timezone.utc)
                completed_dates.add(dt.date())
        if not completed_dates:
            return 0, 0
        # Calculate current streak (ending today)
        streak = 0
        today = datetime.now(timezone.utc).date()
        while True:
            if today in completed_dates:
                streak += 1
                today = today - timedelta(days=1)
            else:
                break
        # Calculate longest streak
        all_dates = sorted(completed_dates)
        longest = 1
        current = 1
        for i in range(1, len(all_dates)):
            if (all_dates[i] - all_dates[i-1]).days == 1:
                current += 1
                if current > longest:
                    longest = current
            else:
                current = 1
        return streak, longest


class FocusSettingsRepository(BaseMongoRepository[FocusSettings]):
    def __init__(self):
        super().__init__(FocusSettings)

    def get_user_settings(self, user_id: str) -> FocusSettings:
        """Get a user's focus settings, creating default settings if none exist"""
        settings = self.find_one({"user_id": user_id})
        if not settings:
            # Create default settings
            settings = FocusSettings(
                user_id=user_id,
                daily_target_seconds=14400,  # 4 hours
                weekly_target_seconds=72000,  # 20 hours
                streak_target_days=5
            )
            self.insert(settings)
            # Refresh settings from database
            new_settings = self.find_one({"user_id": user_id})
            # If we still can't find it, return the original settings object
            if not new_settings:
                return settings
            return new_settings
        return settings

    def update_settings(self, user_id: str, settings_data: dict) -> FocusSettings:
        """Update a user's focus settings"""
        settings = self.find_one({"user_id": user_id})
        if not settings:
            # Create settings with provided data
            new_settings = FocusSettings(
                user_id=user_id,
                **settings_data,
                updated_at=datetime.utcnow()
            )
            self.insert(new_settings)
            # Refresh settings from database
            result = self.find_one({"user_id": user_id})
            # If we still can't find it, return the original settings object
            if not result:
                return new_settings
            return result
        else:
            # Ensure settings has an ID
            if not settings.id:
                # If no ID, recreate the settings
                return self.update_settings(user_id, settings_data)

            # Update existing settings
            settings_data["updated_at"] = datetime.utcnow()

            # Use find_one_and_update pattern
            collection = self.get_collection()
            try:
                obj_id = ObjectId(settings.id)
                result = collection.find_one_and_update(
                    {"_id": obj_id},
                    {"$set": settings_data},
                    return_document=ReturnDocument.AFTER
                )
                if result:
                    return self.model_class.from_mongodb(result)
            except Exception as e:
                logger.error(f"Error updating focus settings: {str(e)}")

            # If update failed, return the original settings
            return settings
