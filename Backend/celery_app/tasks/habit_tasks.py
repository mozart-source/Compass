from Backend.celery_app import celery_app
from Backend.data_layer.repositories.daily_habits_repository import DailyHabitRepository
from Backend.data_layer.database.connection import get_db
import logging
from datetime import datetime
from Backend.celery_app.utils import async_to_sync, task_with_retry

logger = logging.getLogger(__name__)


@celery_app.task(name="habits.process_daily_reset")
# Retry up to 3 times with 5 min delay
@task_with_retry(max_retries=3, countdown=300)
def process_daily_habit_reset():
    """
    Process daily reset operations for habits:
    1. Reset completion status for all habits
    2. Check and reset broken streaks

    This task should be scheduled to run once per day, typically at midnight.
    """
    return async_to_sync(_process_daily_habit_reset)()


async def _process_daily_habit_reset():
    """Async implementation of the daily habit reset process."""
    logger.info(f"Starting daily habit reset process at {datetime.now()}")

    try:
        async for db in get_db():
            repository = DailyHabitRepository(db)

            # Reset all daily completions
            reset_count = await repository.reset_daily_completions()
            logger.info(f"Reset completion status for {reset_count} habits")

            # Check and reset broken streaks
            streak_reset_count = await repository.check_and_reset_broken_streaks()
            logger.info(f"Reset streaks for {streak_reset_count} habits")

            await db.commit()

            return {
                "completions_reset": reset_count,
                "streaks_reset": streak_reset_count,
                "timestamp": datetime.now().isoformat()
            }
    except Exception as e:
        logger.error(f"Error in daily habit reset process: {str(e)}")
        raise
