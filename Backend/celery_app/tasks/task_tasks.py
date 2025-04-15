from Backend.celery_app import celery_app
from Backend.data_layer.repositories.task_repository import TaskRepository
from Backend.data_layer.database.connection import get_db
from Backend.data_layer.database.models.task import Task, TaskStatus
from Backend.celery_app.utils import async_to_sync, task_with_retry
import logging
from typing import Dict, Any, Optional, List
from datetime import datetime

logger = logging.getLogger(__name__)


@celery_app.task(name="tasks.update_status")
@task_with_retry()
def update_task_status(task_id: int, new_status: str, user_id: int) -> Optional[Dict]:
    """Update a task's status asynchronously with validation."""
    return async_to_sync(_update_task_status)(task_id, new_status, user_id)


async def _update_task_status(task_id: int, new_status: str, user_id: int) -> Optional[Dict]:
    """Async implementation of task status update with validation."""
    async for session in get_db():
        try:
            task_repo = TaskRepository(session)
            task = await task_repo.get_task(task_id)
            if not task:
                return None

            # Convert string status to enum if needed
            if isinstance(new_status, str):
                try:
                    new_status = TaskStatus(new_status)
                except ValueError:
                    logger.error(f"Invalid task status: {new_status}")
                    return None

            # Validate status transition
            if not _is_valid_status_transition(task.status, new_status):
                logger.error(
                    f"Invalid status transition from {task.status} to {new_status}")
                return None

            # Update task status
            updated_task = await task_repo.update_task(task_id, {"status": new_status})

            # Add task history entry
            await task_repo.add_task_history(
                task_id=task_id,
                user_id=user_id,
                field_name="status",
                old_value=str(task.status),
                new_value=str(new_status)
            )

            await session.commit()

            # Convert to dict for serialization
            return updated_task.to_dict() if hasattr(updated_task, 'to_dict') else {
                "id": updated_task.id,
                "title": updated_task.title,
                "description": updated_task.description,
                "status": str(updated_task.status),
                "user_id": updated_task.user_id,
                "created_at": updated_task.created_at.isoformat() if updated_task.created_at else None,
                "updated_at": updated_task.updated_at.isoformat() if updated_task.updated_at else None
            }
        except Exception as e:
            await session.rollback()
            logger.error(f"Error updating task status: {str(e)}")
            raise
    return None


def _is_valid_status_transition(current_status: TaskStatus, new_status: TaskStatus) -> bool:
    """Validate if the status transition is allowed."""
    # Define valid transitions
    valid_transitions = {
        TaskStatus.UPCOMING: [TaskStatus.IN_PROGRESS, TaskStatus.BLOCKED, TaskStatus.DEFERRED, TaskStatus.COMPLETED],
        TaskStatus.IN_PROGRESS: [TaskStatus.BLOCKED, TaskStatus.COMPLETED, TaskStatus.DEFERRED],
        TaskStatus.BLOCKED: [TaskStatus.UPCOMING, TaskStatus.IN_PROGRESS],
        TaskStatus.COMPLETED: [TaskStatus.IN_PROGRESS],  # Allow reopening
        TaskStatus.DEFERRED: [TaskStatus.UPCOMING, TaskStatus.IN_PROGRESS]
    }

    return new_status in valid_transitions.get(current_status, [])


@celery_app.task(name="tasks.update_dependencies")
@task_with_retry()
def update_task_dependencies(task_id: int, dependencies: List[int]) -> bool:
    """Update task dependencies asynchronously."""
    return async_to_sync(_update_task_dependencies)(task_id, dependencies)


async def _update_task_dependencies(task_id: int, dependencies: List[int]) -> bool:
    """Async implementation of updating task dependencies."""
    async for session in get_db():
        try:
            task_repo = TaskRepository(session)
            result = await task_repo.update_task_dependencies(task_id, dependencies)
            await session.commit()
            return result
        except Exception as e:
            await session.rollback()
            logger.error(f"Error updating task dependencies: {str(e)}")
            raise
    return False


@celery_app.task(name="tasks.process_task")
@task_with_retry()
async def process_task(task_data: Dict) -> Dict:
    """Main task processing function"""
    try:
        async for session in get_db():
            task_repo = TaskRepository(session)
            # Add actual task processing logic here
            return {"status": "success", "task_id": task_data.get('id')}
    except Exception as e:
        logger.error(f"Task processing failed: {str(e)}")
        raise


@celery_app.task(name="tasks.execute_task_step")
@task_with_retry()
async def execute_task_step(step_data: Dict) -> Dict:
    """Execute individual task steps"""
    try:
        async for session in get_db():
            task_repo = TaskRepository(session)
            # Add step execution logic here
            return {"status": "success", "step": step_data.get('name')}
    except Exception as e:
        logger.error(f"Step execution failed: {str(e)}")
        raise
