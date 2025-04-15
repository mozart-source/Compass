from Backend.orchestration.handlers.base_handler import BaseHandler
from Backend.data_layer.repositories.todo_repository import TodoRepository
from typing import Dict, Any


class TodoHandler(BaseHandler):
    async def enrich_context(self, context: Dict[str, Any]) -> Dict[str, Any]:
        """
        Add high-priority todos to the context.
        """
        todo_repo = TodoRepository(self.db)
        urgent_todos = await todo_repo.get_user_todos(
            user_id=context["user_id"], status="IN_PROGRESS"
        )
        context["urgent_todos"] = [
            {"title": todo.title, "due_date": str(todo.due_date)} for todo in urgent_todos
        ]
        return context
