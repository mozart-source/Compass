from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime


class GoalBase(BaseModel):
    title: str = Field(...)
    description: Optional[str] = None
    target_type: str = Field(...)
    target_value: float = Field(...)
    period: str = Field(...)
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None


class GoalCreate(GoalBase):
    pass


class GoalUpdate(BaseModel):
    title: Optional[str] = None
    description: Optional[str] = None
    target_type: Optional[str] = None
    target_value: Optional[float] = None
    period: Optional[str] = None
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
    progress: Optional[float] = None
    completed: Optional[bool] = None


class GoalResponse(GoalBase):
    id: str
    user_id: str
    progress: float
    completed: bool
    created_at: datetime
    updated_at: datetime


class GoalListResponse(BaseModel):
    goals: List[GoalResponse]
