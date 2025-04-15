from typing import Optional, Dict, Any, List, ClassVar
from pydantic import Field
from data_layer.models.base_model import MongoBaseModel
from datetime import datetime


class FocusSession(MongoBaseModel):
    user_id: str = Field(..., description="User ID")
    start_time: datetime = Field(..., description="Focus session start")
    end_time: Optional[datetime] = Field(None, description="Focus session end")
    duration: Optional[int] = Field(None, description="Duration in seconds")
    status: str = Field("active", description="active|completed|interrupted")
    tags: List[str] = Field(default_factory=list)
    interruptions: int = Field(0, description="Number of interruptions")
    notes: Optional[str] = Field(None)
    metadata: Dict[str, Any] = Field(default_factory=dict)

    collection_name: ClassVar[str] = "focus_sessions"


class FocusSettings(MongoBaseModel):
    user_id: str = Field(..., description="User ID")
    daily_target_seconds: int = Field(
        14400, description="Daily focus target in seconds (default: 4 hours)")
    weekly_target_seconds: int = Field(
        72000, description="Weekly focus target in seconds (default: 20 hours)")
    streak_target_days: int = Field(
        5, description="Target consecutive days for streak achievement")
    created_at: datetime = Field(default_factory=datetime.utcnow)
    updated_at: datetime = Field(default_factory=datetime.utcnow)

    collection_name: ClassVar[str] = "focus_settings"
