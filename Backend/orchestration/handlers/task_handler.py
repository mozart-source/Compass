from Backend.orchestration.handlers.base_handler import BaseHandler
from Backend.data_layer.repositories.task_repository import TaskRepository
from typing import Dict, Any


class TaskHandler(BaseHandler):
    async def enrich_context(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """
        Enhance task context with additional metadata (e.g., recent tasks, priorities).
        """
        task_repo = TaskRepository(self.db)
        recent_tasks = await task_repo.get_recent_tasks(
            user_id=context["user_id"], days=7, limit=5
        )
        context["recent_tasks"] = [
            {"title": task.title, "status": task.status.value} for task in recent_tasks
        ]
        return context
