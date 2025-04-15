from Backend.ai_services.rag.rag_service import RAGService
from Backend.events.event_dispatcher import EventDispatcher
from Backend.events.event_registry import TASK_UPDATED, TODO_UPDATED
from Backend.utils.logging_utils import get_logger

logger = get_logger(__name__)

dispatcher = EventDispatcher()
rag_service = RAGService()

async def on_task_updated(payload):
    task_id = payload.get("task_id")
    logger.info(f"Task updated: {task_id}")
    
    # Retrieve updated task and store it in the knowledge base
    updated_task = payload.get("task_data")
    await rag_service.add_to_knowledge_base(
        content=updated_task["description"],
        metadata={"task_id": task_id, "type": "task"}
    )

dispatcher.register_listener(TASK_UPDATED, on_task_updated)


async def on_todo_updated(payload):
    todo_id = payload.get("todo_id")
    logger.info(f"Todo updated: {todo_id}")
    
    # Retrieve updated todo and store it in the knowledge base
    updated_todo = payload.get("todo_data")
    await rag_service.add_to_knowledge_base(
        content=updated_todo["description"],
        metadata={"todo_id": todo_id, "type": "todo"}
    )
    
dispatcher.register_listener(TODO_UPDATED, on_todo_updated)