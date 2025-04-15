from typing import Dict, Any, Optional, TypeVar, Generic, List, Union, cast
from crewai import Task, Agent
from datetime import datetime

T = TypeVar('T', bound=Dict[str, Any])


class CompassTask(Task):
    """Custom Task class for COMPASS project that handles task context properly."""

    def __init__(
        self,
        description: str,
        agent: Agent,
        expected_output: str,
        task_data: Dict[str, Any],
        task_type: Optional[str] = None
    ):
        """Initialize a CompassTask with proper context handling."""
        # Initialize the parent Task class first
        super().__init__(
            description=description,
            agent=agent,
            expected_output=expected_output
        )
        
        # Set task-specific attributes after parent initialization
        self._task_data = task_data
        self._task_type = task_type or "general"

    def get_task_data(self) -> Dict[str, Any]:
        """Get the task data."""
        return self._task_data

    def get_task_type(self) -> str:
        """Get the task type."""
        return self._task_type

    def update_task_data(self, new_data: Dict[str, Any]) -> None:
        """Update task data with new information."""
        if isinstance(self._task_data, dict):
            self._task_data.update(new_data)

    def to_dict(self) -> Dict[str, Any]:
        """Convert task to dictionary format."""
        return {
            "description": self.description,
            "expected_output": self.expected_output,
            "task_type": self.task_type,
            "task_data": self.task_data
        }

    @classmethod
    def from_db_task(cls, db_task: Any, agent: Agent, expected_output: str = "Task completion") -> "CompassTask":
        """Create a CompassTask from a database Task model.

        Args:
            db_task: Database Task model instance
            agent: Agent to assign to the task
            expected_output: Expected output description

        Returns:
            CompassTask instance with data from the database model
        """
        # Extract relevant data from the database model
        task_data = {
            "task_id": db_task.id,
            "title": db_task.title,
            "description": db_task.description,
            "status": str(db_task.status.value) if hasattr(db_task.status, "value") else str(db_task.status),
            "priority": str(db_task.priority.value) if hasattr(db_task.priority, "value") else str(db_task.priority),
            "due_date": db_task.due_date.isoformat() if db_task.due_date else None,
            "estimated_hours": db_task.estimated_hours,
            "dependencies": db_task.dependencies if hasattr(db_task, "dependencies") else [],
            "blockers": db_task.blockers if hasattr(db_task, "blockers") else [],
            "workflow_id": db_task.workflow_id
        }

        # Add AI-specific fields if they exist
        if hasattr(db_task, "ai_suggestions") and db_task.ai_suggestions:
            task_data["ai_suggestions"] = db_task.ai_suggestions

        if hasattr(db_task, "complexity_score") and db_task.complexity_score:
            task_data["complexity_score"] = db_task.complexity_score

        return cls(
            description=f"Process task: {db_task.title}",
            agent=agent,
            expected_output=expected_output,
            task_data=task_data,
            task_type="db_task"
        )

    def to_db_task_update(self) -> Dict[str, Any]:
        """Convert CompassTask to a dictionary suitable for updating a database Task.

        Returns:
            Dictionary with fields that can be used to update a database Task
        """
        update_data = {}

        if "title" in self.task_data:
            update_data["title"] = self.task_data["title"]

        if "description" in self.task_data:
            update_data["description"] = self.task_data["description"]

        if "status" in self.task_data:
            update_data["status"] = self.task_data["status"]

        if "priority" in self.task_data:
            update_data["priority"] = self.task_data["priority"]

        if "due_date" in self.task_data and self.task_data["due_date"]:
            # Handle string ISO format conversion to datetime if needed
            if isinstance(self.task_data["due_date"], str):
                update_data["due_date"] = datetime.fromisoformat(
                    self.task_data["due_date"].replace('Z', '+00:00'))
            else:
                update_data["due_date"] = self.task_data["due_date"]

        if "estimated_hours" in self.task_data:
            update_data["estimated_hours"] = self.task_data["estimated_hours"]

        if "dependencies" in self.task_data:
            update_data["dependencies"] = self.task_data["dependencies"]

        if "blockers" in self.task_data:
            update_data["blockers"] = self.task_data["blockers"]

        if "ai_suggestions" in self.task_data:
            update_data["ai_suggestions"] = self.task_data["ai_suggestions"]

        if "complexity_score" in self.task_data:
            update_data["complexity_score"] = self.task_data["complexity_score"]

        return update_data
