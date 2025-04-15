from typing import Dict, List, Optional, Any
from datetime import datetime, date, timedelta
from crewai import Agent, Task, Crew
from langchain.tools import Tool
from Backend.ai_services.llm.llm_service import LLMService
from Backend.utils.logging_utils import get_logger
from pydantic import Field
from Backend.services.task_service import TaskService
from Backend.data_layer.database.models.task import TaskStatus, TaskPriority
from Backend.data_layer.database.models.calendar_event import RecurrenceType
import json
import re
from sqlalchemy.ext.asyncio import AsyncSession

logger = get_logger(__name__)


class TaskCreationAgent:
    def __init__(self, db_session=None):
        self.agent = Agent(
            role="Task Creation Specialist",
            goal="Create tasks from natural language instructions",
            backstory="""I am an expert in task management and organization. I help users 
                        efficiently create and manage their tasks by understanding natural 
                        language and extracting structured information.""",
            tools=[
                Tool(
                    name="create_task",
                    func=self.create_task,
                    description="Creates a new task from natural language"
                )
            ],
            verbose=True,
            allow_delegation=True  # Enable task delegation
        )
        
        self.ai_service = LLMService()
        self.db = db_session
        
        # Initialize services instead of repositories
        if db_session:
            # For TaskService, we need to provide a TaskRepository instance
            from Backend.data_layer.repositories.task_repository import TaskRepository
            task_repo = TaskRepository(db_session)
            self.task_service = TaskService(task_repo)
        else:
            self.task_service = None
    
    async def create_task(self, description: str, user_id: int = None) -> Dict[str, Any]:
        """Create a new task from natural language description."""
        try:
            # Set current_time once at the beginning
            current_time = datetime.utcnow()
            
            # Use LLM to extract structured data from description
            task_analysis = await self.ai_service.generate_response(
                prompt=f"""Extract structured task information from this description:
Description: {description}

Extract and format as JSON:
- title: A clear, concise title for the task
- description: Full description with details
- priority: high, medium, or low
- due_date: Extract any mentioned deadline (YYYY-MM-DD format)
- estimated_hours: Numerical estimate of hours needed (if mentioned)
- tags: Array of relevant tags
- status: Default is "todo", other options are "in_progress", "completed", "cancelled", "blocked", "under_review", "deferred" """,
                context={
                    "system_message": "You are a task extraction AI. Extract structured task information from natural language."
                }
            )
            
            try:
                # Parse the JSON response
                response_text = task_analysis.get("text", "")
                
                # Extract JSON from markdown if needed
                if "```json" in response_text:
                    json_match = re.search(r'```json\n(.*?)\n```', response_text, re.DOTALL)
                    if json_match:
                        response_text = json_match.group(1)
                
                task_data = json.loads(response_text)
                
                # Map status string to proper TaskStatus enum value
                # Map common status terms to our enum values
                status_mapping = {
                    "pending": "TODO",
                    "todo": "TODO",
                    "to do": "TODO",
                    "to-do": "TODO",
                    "in progress": "IN_PROGRESS",
                    "in-progress": "IN_PROGRESS",
                    "inprogress": "IN_PROGRESS",
                    "done": "COMPLETED",
                    "complete": "COMPLETED",
                    "completed": "COMPLETED",
                    "cancel": "CANCELLED",
                    "canceled": "CANCELLED",
                    "cancelled": "CANCELLED",
                    "block": "BLOCKED",
                    "blocked": "BLOCKED",
                    "review": "UNDER_REVIEW",
                    "under review": "UNDER_REVIEW",
                    "under-review": "UNDER_REVIEW",
                    "defer": "DEFERRED",
                    "deferred": "DEFERRED"
                }
                
                if "status" in task_data:
                    status_str = task_data["status"].lower()
                    if status_str in status_mapping:
                        enum_value = status_mapping[status_str]
                        try:
                            task_data["status"] = TaskStatus[enum_value]
                        except KeyError:
                            logger.warning(f"Invalid status mapping: {enum_value}, using TODO")
                            task_data["status"] = TaskStatus.UPCOMING
                    else:
                        # Try direct enum lookup
                        try:
                            task_data["status"] = TaskStatus[status_str.upper()]
                        except KeyError:
                            logger.warning(f"Invalid status: {status_str}, using TODO")
                            task_data["status"] = TaskStatus.UPCOMING
                else:
                    task_data["status"] = TaskStatus.UPCOMING
                
                # Convert priority string to enum value
                priority_mapping = {
                    "low": "LOW",
                    "medium": "MEDIUM",
                    "mid": "MEDIUM",
                    "high": "HIGH",
                    "urgent": "URGENT",
                    "critical": "URGENT"
                }
                
                if "priority" in task_data:
                    priority_str = task_data["priority"].lower()
                    if priority_str in priority_mapping:
                        enum_value = priority_mapping[priority_str]
                        try:
                            task_data["priority"] = TaskPriority[enum_value]
                        except KeyError:
                            logger.warning(f"Invalid priority mapping: {enum_value}, using MEDIUM")
                            task_data["priority"] = TaskPriority.MEDIUM
                    else:
                        # Try direct enum lookup
                        try:
                            task_data["priority"] = TaskPriority[priority_str.upper()]
                        except KeyError:
                            logger.warning(f"Invalid priority: {priority_str}, using MEDIUM")
                            task_data["priority"] = TaskPriority.MEDIUM
                else:
                    task_data["priority"] = TaskPriority.MEDIUM
                
                # Set recurrence type to NONE by default
                task_data["recurrence"] = RecurrenceType.NONE
                
                # Convert due_date string to datetime object if present
                if "due_date" in task_data and task_data["due_date"]:
                    try:
                        # Parse YYYY-MM-DD format
                        due_date = datetime.strptime(task_data["due_date"], "%Y-%m-%d")
                        
                        # Ensure due_date is not in the past
                        if due_date < current_time:
                            # If due date is in the past, set it to end of current day
                            logger.warning(f"Due date {due_date} is in the past, adjusting to future date")
                            adjusted_due_date = current_time.replace(hour=23, minute=59, second=59)
                            task_data["due_date"] = adjusted_due_date
                        else:
                            task_data["due_date"] = due_date
                    except ValueError:
                        # If parsing fails, remove invalid date
                        logger.warning(f"Invalid due_date format: {task_data['due_date']}")
                        task_data.pop("due_date", None)
                
                # Map task_data to parameters needed for TaskService.create_task
                task_creation_data = {
                    "title": task_data["title"],
                    "description": task_data.get("description", ""),
                    "creator_id": user_id,
                    "organization_id": 1,  # Default organization_id
                    "project_id": 1,  # Default project_id
                    "start_date": current_time,
                    "status": task_data["status"],
                    "priority": task_data["priority"],
                    "recurrence": task_data["recurrence"],
                    "due_date": task_data.get("due_date"),
                    # Safely handle None values for duration
                    "duration": float(task_data.get("estimated_hours")) if task_data.get("estimated_hours") is not None else 1.0,
                    # Safely handle None values for estimated_hours
                    "estimated_hours": float(task_data.get("estimated_hours")) if task_data.get("estimated_hours") is not None else None,
                    "dependencies": None  # Can be added if extracted from description
                }
                
                # Save to database if service is available
                if self.task_service is not None:
                    db_task = await self.task_service.create_task(**task_creation_data)
                    task_data["id"] = db_task.id
                
                # Ensure all values are JSON serializable
                # Convert datetime objects to strings
                for key, value in list(task_data.items()):
                    if isinstance(value, datetime):
                        task_data[key] = value.isoformat()
                    elif hasattr(value, "name"):  # Check if it's an enum
                        task_data[key] = value.name
                    elif isinstance(value, (list, tuple)):
                        # Handle lists that might contain non-serializable items
                        task_data[key] = [
                            item.name if hasattr(item, "name") else 
                            item.isoformat() if hasattr(item, "isoformat") else 
                            item 
                            for item in value
                        ]
                
                return {
                    "status": "success",
                    "message": f"Task '{task_data['title']}' created successfully",
                    "task": task_data
                }
            except Exception as e:
                logger.error(f"Error parsing task data: {str(e)}")
                return {
                    "status": "error",
                    "message": f"Could not parse task information: {str(e)}",
                    "raw_response": task_analysis.get("text", "")
                }
        except Exception as e:
            logger.error(f"Task creation failed: {str(e)}")
            return {"status": "error", "message": f"Task creation failed: {str(e)}"} 