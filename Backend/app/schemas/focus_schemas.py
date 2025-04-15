from pydantic import BaseModel, Field
from typing import Optional, List
from datetime import datetime


class FocusSessionCreate(BaseModel):
    start_time: datetime
    tags: Optional[List[str]] = []
    notes: Optional[str] = None


class FocusSessionStop(BaseModel):
    end_time: datetime
    notes: Optional[str] = None


class FocusSessionResponse(BaseModel):
    id: str
    user_id: str
    start_time: datetime
    end_time: Optional[datetime]
    duration: Optional[int]
    status: str
    tags: List[str]
    interruptions: int
    notes: Optional[str]


class FocusStatsResponse(BaseModel):
    total_focus_seconds: int
    streak: int
    longest_streak: int
    sessions: int
    daily_target_seconds: Optional[int] = None


class FocusSettingsUpdate(BaseModel):
    daily_target_seconds: Optional[int] = None
    weekly_target_seconds: Optional[int] = None
    streak_target_days: Optional[int] = None


class FocusSettingsResponse(BaseModel):
    id: str
    user_id: str
    daily_target_seconds: int
    weekly_target_seconds: int
    streak_target_days: int
    created_at: datetime
    updated_at: datetime
