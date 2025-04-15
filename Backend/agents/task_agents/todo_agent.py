from typing import Dict, List, Optional, Any
from datetime import datetime, date, timedelta
from crewai import Agent, Task, Crew
from langchain.tools import Tool
from Backend.ai_services.llm.llm_service import LLMService
from Backend.utils.logging_utils import get_logger
from pydantic import Field
from Backend.services.todo_service import TodoService
from Backend.data_layer.database.models.todo import TodoStatus, TodoPriority
import json
import re
from sqlalchemy.ext.asyncio import AsyncSession

logger = get_logger(__name__)


class TodoAgent:
    def __init__(self, db_session=None):
        self.agent = Agent(
            role="Todo Management Specialist",
            goal="Create and edit todos from natural language instructions",
            backstory="""I am an expert in todo management and organization. I help users 
                        efficiently create and manage their todos by understanding natural 
                        language and extracting structured information.""",
            tools=[
                Tool(
                    name="create_todo",
                    func=self.create_todo,
                    description="Creates a new todo item from natural language"
                ),
                Tool(
                    name="edit_todo",
                    func=self.edit_todo,
                    description="Processes a natural language request to edit a todo, identifying which todo to edit and how"
                )
            ],
            verbose=True,
            allow_delegation=True  # Enable task delegation
        )
        
        self.ai_service = LLMService()
        self.db = db_session
        
        # Initialize services instead of repositories
        if db_session:
            # For TodoService, we need to provide a TodoRepository instance
            from Backend.data_layer.repositories.todo_repository import TodoRepository
            todo_repo = TodoRepository(db_session)
            self.todo_service = TodoService(todo_repo)
        else:
            self.todo_service = None
    
    async def create_todo(self, description: str, user_id: int = None) -> Dict[str, Any]:
        """Create a new todo item from natural language description."""
        try:
            # Set current_time once at the beginning
            current_time = datetime.utcnow()
            
            # Use LLM to extract structured data
            todo_analysis = await self.ai_service.generate_response(
                prompt=f"""Extract structured todo information from this description:
Description: {description}

Extract and format as JSON:
- title: A clear, concise title for the todo item
- description: Any additional details (can be empty)
- priority: high, medium, or low
- due_date: Extract any mentioned deadline (YYYY-MM-DD format)
- tags: Array of relevant tags""",
                context={
                    "system_message": "You are a todo extraction AI. Extract structured todo information from natural language."
                }
            )
            
            try:
                # Parse the JSON response
                response_text = todo_analysis.get("text", "")
                
                # Extract JSON from markdown if needed
                if "```json" in response_text:
                    json_match = re.search(r'```json\n(.*?)\n```', response_text, re.DOTALL)
                    if json_match:
                        response_text = json_match.group(1)
                
                todo_data = json.loads(response_text)
                
                # Add metadata - use actual datetime objects, not strings
                todo_data["created_at"] = current_time
                todo_data["updated_at"] = current_time
                todo_data["is_recurring"] = False
                todo_data["user_id"] = user_id
                
                # Set status to PENDING by default
                todo_data["status"] = TodoStatus.PENDING
                
                # Convert priority string to enum value
                priority_mapping = {
                    "low": TodoPriority.LOW,
                    "medium": TodoPriority.MEDIUM,
                    "mid": TodoPriority.MEDIUM,
                    "high": TodoPriority.HIGH
                }
                
                if "priority" in todo_data:
                    priority_str = todo_data["priority"].lower()
                    todo_data["priority"] = priority_mapping.get(priority_str, TodoPriority.MEDIUM)
                else:
                    todo_data["priority"] = TodoPriority.MEDIUM
                
                # Convert due_date string to datetime object if present
                if "due_date" in todo_data and todo_data["due_date"]:
                    try:
                        # Parse YYYY-MM-DD format
                        due_date = datetime.strptime(todo_data["due_date"], "%Y-%m-%d")
                        
                        # Ensure due_date is not in the past
                        if due_date < current_time:
                            # If due date is in the past, set it to end of current day
                            logger.warning(f"Due date {due_date} is in the past, adjusting to future date")
                            adjusted_due_date = current_time.replace(hour=23, minute=59, second=59)
                            todo_data["due_date"] = adjusted_due_date
                        else:
                            todo_data["due_date"] = due_date
                    except ValueError:
                        # If parsing fails, remove invalid date
                        logger.warning(f"Invalid due_date format: {todo_data['due_date']}")
                        todo_data.pop("due_date", None)
                
                # Save to database if service is available
                if self.todo_service is not None:
                    db_todo = await self.todo_service.create_todo(**todo_data)
                    if db_todo:
                        todo_data["id"] = db_todo.id if hasattr(db_todo, 'id') else None
                
                # Ensure all values are JSON serializable
                # Convert datetime objects to strings
                for key, value in list(todo_data.items()):
                    if isinstance(value, datetime):
                        todo_data[key] = value.isoformat()
                    elif isinstance(value, (TodoStatus, TodoPriority)):  # Check if it's an enum
                        todo_data[key] = value.value
                    elif isinstance(value, (list, tuple)):
                        # Handle lists that might contain non-serializable items
                        todo_data[key] = [
                            item.value if isinstance(item, (TodoStatus, TodoPriority)) else 
                            item.isoformat() if hasattr(item, "isoformat") else 
                            item 
                            for item in value
                        ]
                
                return {
                    "status": "success",
                    "message": f"Todo '{todo_data['title']}' created successfully",
                    "todo": todo_data
                }
            except Exception as e:
                logger.error(f"Error parsing todo data: {str(e)}")
                return {
                    "status": "error",
                    "message": f"Could not parse todo information: {str(e)}",
                    "raw_response": todo_analysis.get("text", "")
                }
        except Exception as e:
            logger.error(f"Todo creation failed: {str(e)}")
            return {"status": "error", "message": f"Todo creation failed: {str(e)}"}

    async def edit_todo(self, description: str, user_id: int = None, previous_messages: List[Dict[str, str]] = None) -> Dict[str, Any]:
        """Edit an existing todo item based on natural language description."""
        try:
            current_time = datetime.utcnow()
            
            # Prepare context for identifying which todo to edit
            context_prompt = ""
            
            # Add previous conversation messages for context if available
            if previous_messages and len(previous_messages) > 0:
                context_prompt += "Previous conversation:\n"
                # Add up to 5 most recent messages for context
                for msg in previous_messages[-5:]:
                    role = msg.get("role", "unknown")
                    content = msg.get("content", "")
                    metadata = msg.get("metadata", {})
                    timestamp = msg.get("timestamp")
                    
                    # Format timestamp if available
                    time_str = ""
                    if timestamp:
                        try:
                            time_str = f" ({datetime.fromtimestamp(timestamp).strftime('%H:%M:%S')})"
                        except:
                            pass
                    
                    context_prompt += f"{role.capitalize()}{time_str}: {content}\n"
                    
                    # Add any relevant metadata
                    if metadata and metadata.get("domain"):
                        context_prompt += f"Context: {metadata['domain']}\n"
                context_prompt += "\n"
            
            # Get user's todos for context if todo_service is available
            todos_context = ""
            if self.todo_service is not None:
                try:
                    # Get all todos for this user
                    user_todos = await self.todo_service.get_user_todos(user_id=user_id)
                    if user_todos:
                        todos_context += "User's current todos:\n"
                        for i, todo in enumerate(user_todos, 1):
                            # Format each todo with key information
                            # Handle both dict and object formats
                            if isinstance(todo, dict):
                                todo_status = todo.get("status")
                                todo_priority = todo.get("priority")
                                todo_due_date = todo.get("due_date", "No due date")
                                todo_id = todo.get("id")
                                todo_title = todo.get("title")
                            else:
                                # Original object access
                                todo_status = todo.status.value if hasattr(todo.status, 'value') else todo.status
                                todo_priority = todo.priority.value if hasattr(todo.priority, 'value') else todo.priority
                                todo_due_date = todo.due_date.strftime("%Y-%m-%d") if todo.due_date else "No due date"
                                todo_id = todo.id
                                todo_title = todo.title
                            
                            todos_context += f"{i}. ID: {todo_id} - Title: '{todo_title}' - Status: {todo_status} - "
                            todos_context += f"Priority: {todo_priority} - Due: {todo_due_date}\n"
                except Exception as e:
                    logger.error(f"Error fetching user todos: {str(e)}")
                    todos_context = "Error fetching user todos\n"
            
            # Use LLM to identify which todo to edit and what changes to make
            edit_analysis = await self.ai_service.generate_response(
                prompt=f"""Identify which todo to edit and what changes to make based on this user request:

{context_prompt}
{todos_context}

User request: {description}

Return a STRICT JSON object with the following structure (no additional text):
{{
    "todo_id": number,
    "changes": {{
        "title"?: string,
        "description"?: string,
        "status"?: "PENDING" | "IN_PROGRESS" | "COMPLETED" | "ARCHIVED",
        "priority"?: "LOW" | "MEDIUM" | "HIGH",
        "due_date"?: "YYYY-MM-DD",
        "tags"?: string[]
    }},
    "reason": string
}}

Note: Only include fields in "changes" that need to be updated. Use EXACTLY the enum values shown above for status and priority.""",
                context={
                    "system_message": "You are a todo management AI. Generate valid JSON for todo updates."
                }
            )
            
            try:
                # Parse the JSON response
                response_text = edit_analysis.get("text", "")
                logger.debug(f"Raw LLM response:\n{response_text}")
                
                # Clean and parse JSON
                response_text = (response_text
                    .strip()                     # Remove leading/trailing whitespace
                    .replace('\n', ' ')          # Remove newlines
                    .replace('\r', '')           # Remove carriage returns
                )
                
                # Remove any markdown code block markers if present
                response_text = response_text.replace('```json', '').replace('```', '')
                
                logger.debug(f"After cleanup:\n{response_text}")
                try:
                    edit_data = json.loads(response_text)
                except json.JSONDecodeError as e:
                    logger.error(f"JSON parse error at position {e.pos}. Near text: {response_text[max(0, e.pos-50):min(len(response_text), e.pos+50)]}")
                    raise
                
                # Ensure we have the required todo_id
                if "todo_id" not in edit_data or not edit_data["todo_id"]:
                    return {
                        "status": "error",
                        "message": "Could not identify which todo to edit",
                        "raw_response": response_text
                    }
                
                # Ensure todo_id is an integer
                try:
                    # Convert string ID to integer if needed
                    todo_id = int(str(edit_data["todo_id"]).strip())
                    edit_data["todo_id"] = todo_id
                except (ValueError, TypeError):
                    return {
                        "status": "error", 
                        "message": f"Invalid todo ID format: {edit_data['todo_id']}. Expected a numeric ID.",
                        "raw_response": response_text
                    }
                
                # Ensure we have changes to apply
                if "changes" not in edit_data or not edit_data["changes"]:
                    return {
                        "status": "error",
                        "message": "No changes specified",
                        "raw_response": response_text
                    }
                
                # Prepare update data
                update_data = edit_data["changes"]
                update_data["updated_at"] = current_time
                
                # Convert status string to enum value if present
                if "status" in update_data:
                    status_mapping = {
                        "pending": TodoStatus.PENDING,
                        "in_progress": TodoStatus.IN_PROGRESS,
                        "completed": TodoStatus.COMPLETED,
                        "archived": TodoStatus.ARCHIVED
                    }
                    status_str = update_data["status"].lower()
                    update_data["status"] = status_mapping.get(status_str, TodoStatus.PENDING)
                
                # Convert priority string to enum value if present
                if "priority" in update_data:
                    priority_mapping = {
                        "low": TodoPriority.LOW,
                        "medium": TodoPriority.MEDIUM,
                        "high": TodoPriority.HIGH
                    }
                    priority_str = update_data["priority"].lower()
                    update_data["priority"] = priority_mapping.get(priority_str, TodoPriority.MEDIUM)
                
                # Convert due_date string to datetime object if present
                if "due_date" in update_data and update_data["due_date"]:
                    try:
                        update_data["due_date"] = datetime.strptime(update_data["due_date"], "%Y-%m-%d")
                    except ValueError:
                        logger.warning(f"Invalid due_date format: {update_data['due_date']}")
                        update_data.pop("due_date", None)
                
                # Apply the update if service is available
                if self.todo_service is not None:
                    todo_id = edit_data["todo_id"]
                    updated_todo = await self.todo_service.update_todo(todo_id, user_id, **update_data)
                    
                    if updated_todo:
                        # Convert the updated todo to a serializable format
                        todo_dict = {
                            "id": updated_todo["id"] if isinstance(updated_todo, dict) else updated_todo.id,
                            "title": updated_todo["title"] if isinstance(updated_todo, dict) else updated_todo.title,
                            "description": updated_todo["description"] if isinstance(updated_todo, dict) else updated_todo.description,
                            "status": updated_todo["status"] if isinstance(updated_todo, dict) else (updated_todo.status.value if hasattr(updated_todo.status, 'value') else updated_todo.status),
                            "priority": updated_todo["priority"] if isinstance(updated_todo, dict) else (updated_todo.priority.value if hasattr(updated_todo.priority, 'value') else updated_todo.priority),
                            "due_date": updated_todo["due_date"] if isinstance(updated_todo, dict) else (updated_todo.due_date.isoformat() if updated_todo.due_date else None),
                            "user_id": updated_todo["user_id"] if isinstance(updated_todo, dict) else updated_todo.user_id,
                            "created_at": updated_todo["created_at"] if isinstance(updated_todo, dict) else updated_todo.created_at.isoformat(),
                            "updated_at": updated_todo["updated_at"] if isinstance(updated_todo, dict) else updated_todo.updated_at.isoformat()
                        }
                        
                        # Handle tags if present
                        if isinstance(updated_todo, dict):
                            todo_dict["tags"] = updated_todo.get("tags", [])
                        else:
                            todo_dict["tags"] = updated_todo.tags if hasattr(updated_todo, 'tags') else []
                        
                        status_change = ""
                        if "status" in update_data:
                            status_change = f" and marked as {update_data['status'].value if hasattr(update_data['status'], 'value') else update_data['status']}"
                        
                        return {
                            "status": "success",
                            "message": f"Todo '{todo_dict['title']}' updated successfully{status_change}",
                            "todo": todo_dict,
                            "changes_applied": list(update_data.keys())
                        }
                    else:
                        return {
                            "status": "error",
                            "message": f"Todo with ID {todo_id} not found or could not be updated",
                            "todo_id": todo_id
                        }
                else:
                    # For testing without database
                    return {
                        "status": "success",
                        "message": "Todo would be updated (no database connection)",
                        "todo_id": edit_data["todo_id"],
                        "changes": update_data,
                        "reason": edit_data.get("reason", "")
                    }
                
            except Exception as e:
                logger.error(f"Error parsing edit data: {str(e)}")
                return {
                    "status": "error",
                    "message": f"Could not parse todo edit information: {str(e)}",
                    "raw_response": response_text
                }
                
        except Exception as e:
            logger.error(f"Todo edit failed: {str(e)}")
            return {"status": "error", "message": f"Todo edit failed: {str(e)}"}