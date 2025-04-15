from typing import Optional, ClassVar
from datetime import datetime
from pydantic import Field
from .base_model import MongoBaseModel


class Goal(MongoBaseModel):
    user_id: str = Field(..., description="User ID")
    title: str = Field(..., description="Goal title")
    description: Optional[str] = Field(None, description="Goal description")
    target_type: str = Field(
        ..., description="Type of target (journals, notes, tasks, focus, mood, custom)")
    target_value: float = Field(
        ..., description="Target value (number or string, e.g., 5 journals)")
    period: str = Field(...,
                        description="Goal period (daily, weekly, monthly, custom)")
    start_date: Optional[datetime] = Field(None, description="Start date")
    end_date: Optional[datetime] = Field(None, description="End date")
    progress: float = Field(0, description="Current progress")
    completed: bool = Field(False, description="Is the goal completed?")
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)

    collection_name: ClassVar[str] = "user_goals"
