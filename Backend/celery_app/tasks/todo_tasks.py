from Backend.celery_app import celery_app
from Backend.data_layer.repositories.todo_repository import TodoRepository
from Backend.data_layer.database.connection import get_db
from Backend.data_layer.database.models.todo import Todo
from Backend.celery_app.utils import async_to_sync, task_with_retry
import logging
from typing import Dict, Any, Optional, List

logger = logging.getLogger(__name__)


@celery_app.task(name="todos.create")
async def create_todo_task(todo_data: Dict[str, Any]) -> Optional[Dict]:
    return await _create_todo(todo_data)


@celery_app.task(name="todos.update")
async def update_todo_task(todo_id: int, user_id: int, updates: Dict[str, Any]) -> Optional[Dict]:
    return await _update_todo(todo_id, user_id, updates)


@celery_app.task(name="todos.delete")
async def delete_todo_task(todo_id: int, user_id: int) -> bool:
    return await _delete_todo(todo_id, user_id)


async def _create_todo(todo_data: Dict[str, Any]) -> Optional[Dict]:
    """Async implementation of todo creation."""
    async for session in get_db():
        try:
            todo_repo = TodoRepository(session)
            todo = await todo_repo.create(**todo_data)
            await session.commit()
            
            # Convert to dict for serialization
            todo_dict = todo.to_dict() if hasattr(todo, 'to_dict') else {
                "id": todo.id,
                "title": todo.title,
                "description": todo.description,
                "status": todo.status,
                "user_id": todo.user_id,
                "created_at": todo.created_at.isoformat() if todo.created_at else None,
                "updated_at": todo.updated_at.isoformat() if todo.updated_at else None
            }
            
            return todo_dict
        except Exception as e:
            await session.rollback()
            logger.error(f"Error creating todo: {str(e)}")
            raise
    return None


async def _update_todo(todo_id: int, user_id: int, updates: Dict[str, Any]) -> Optional[Dict]:
    """Async implementation of todo update."""
    async for session in get_db():
        try:
            todo_repo = TodoRepository(session)
            todo = await todo_repo.get_by_id(todo_id, user_id)
            if not todo:
                return None
            result = await todo_repo.update(todo_id, user_id, **updates)
            await session.commit()
            
            # Convert to dict for serialization
            result_dict = result.to_dict() if hasattr(result, 'to_dict') else {
                "id": result.id,
                "title": result.title,
                "description": result.description,
                "status": result.status,
                "user_id": result.user_id,
                "created_at": result.created_at.isoformat() if result.created_at else None,
                "updated_at": result.updated_at.isoformat() if result.updated_at else None
            }
            
            return result_dict
        except Exception as e:
            await session.rollback()
            logger.error(f"Error updating todo: {str(e)}")
            raise
    return None


async def _delete_todo(todo_id: int, user_id: int) -> bool:
    """Async implementation of todo deletion."""
    async for session in get_db():
        try:
            todo_repo = TodoRepository(session)
            todo = await todo_repo.get_by_id(todo_id, user_id)
            if not todo:
                return False
            result = await todo_repo.delete(todo_id, user_id)
            await session.commit()
            
            return bool(result)
        except Exception as e:
            await session.rollback()
            logger.error(f"Error deleting todo: {str(e)}")
            raise
    return False


@celery_app.task(name="tasks.get_todo_by_id")
async def get_todo_by_id_task(todo_id: int, user_id: int):
    return await _get_todo_by_id(todo_id, user_id)


async def _get_todo_by_id(todo_id: int, user_id: int) -> Optional[Dict]:
    """Async implementation of getting a todo by ID."""
    async for session in get_db():
        try:
            todo_repo = TodoRepository(session)
            todo = await todo_repo.get_by_id(todo_id, user_id)
            if not todo:
                return None
            return todo.to_dict() if hasattr(todo, 'to_dict') else {
                "id": todo.id,
                "title": todo.title,
                "description": todo.description,
                "status": todo.status,
                "user_id": todo.user_id,
                "created_at": todo.created_at.isoformat() if todo.created_at else None,
                "updated_at": todo.updated_at.isoformat() if todo.updated_at else None
            }
        except Exception as e:
            logger.error(f"Error getting todo: {str(e)}")
            raise
    return None


@celery_app.task(name="todos.get_all")
async def get_todos_task(user_id: int, status: Optional[str] = None) -> List[Dict]:
    return await _get_todos(user_id, status)


async def _get_todos(user_id: int, status: Optional[str] = None) -> List[Dict]:
    """Async implementation of getting all todos for a user."""
    async for session in get_db():
        try:
            todo_repo = TodoRepository(session)
            todos = await todo_repo.get_user_todos(user_id, status)
            return [
                todo.to_dict() if hasattr(todo, 'to_dict') else {
                    "id": todo.id,
                    "title": todo.title,
                    "description": todo.description,
                    "status": todo.status,
                    "user_id": todo.user_id,
                    "created_at": todo.created_at.isoformat() if todo.created_at else None,
                    "updated_at": todo.updated_at.isoformat() if todo.updated_at else None
                }
                for todo in todos
            ]
        except Exception as e:
            logger.error(f"Error getting todos: {str(e)}")
            raise
    return []
