from typing import Dict, Any, Optional, ClassVar, List
from pydantic import Field
from data_layer.models.base_model import MongoBaseModel
from datetime import datetime
from enum import Enum


class ReportStatus(str, Enum):
    PENDING = "pending"
    GENERATING = "generating"
    COMPLETED = "completed"
    FAILED = "failed"


class ReportType(str, Enum):
    ACTIVITY = "activity"
    PRODUCTIVITY = "productivity"
    HABITS = "habits"
    TASK = "task"
    SUMMARY = "summary"
    DASHBOARD = "dashboard"
    CUSTOM = "custom"


class Report(MongoBaseModel):
    """Model for AI-generated reports."""

    title: str = Field(..., description="Title of the report")
    user_id: str = Field(..., description="ID of the user who owns the report")
    type: ReportType = Field(..., description="Type of report")
    status: ReportStatus = Field(
        default=ReportStatus.PENDING, description="Current status of report generation")
    parameters: Dict[str, Any] = Field(
        default_factory=dict,
        description="Parameters used to generate the report"
    )
    time_range: Dict[str, str] = Field(
        default_factory=dict,
        description="Time range for the report data (start_date, end_date)"
    )
    custom_prompt: Optional[str] = Field(
        None, description="Custom prompt for report generation")
    content: Optional[Dict[str, Any]] = Field(
        None, description="Generated report content")
    summary: Optional[str] = Field(
        None, description="Brief summary of the report")
    created_at: datetime = Field(
        default_factory=datetime.utcnow, description="When the report was created")
    completed_at: Optional[datetime] = Field(
        None, description="When the report generation was completed")
    sections: List[Dict[str, Any]] = Field(
        default_factory=list,
        description="Report sections and their content"
    )
    error: Optional[str] = Field(
        None, description="Error message if report generation failed")
    model_id: Optional[str] = Field(
        None, description="ID of the AI model used for generation")
    token_usage: Optional[Dict[str, int]] = Field(
        None, description="Token usage statistics for report generation")

    # Set collection name
    collection_name: ClassVar[str] = "reports"

    class Config:
        """Pydantic model configuration."""
        json_encoders = {
            datetime: lambda dt: dt.isoformat()
        }
        json_schema_extra = {
            "example": {
                "title": "Weekly Productivity Report",
                "user_id": "user123",
                "type": "productivity",
                "status": "completed",
                "parameters": {
                    "include_charts": True,
                    "detailed_breakdown": True
                },
                "time_range": {
                    "start_date": "2024-03-15",
                    "end_date": "2024-03-21"
                },
                "content": {
                    "productivity_score": 87,
                    "focus_time": 34.5,
                    "completed_tasks": 28,
                    "recommendations": ["Morning focus blocks", "Reduce context switching"]
                },
                "summary": "Overall productivity is up 12% from last week with improved focus time.",
                "created_at": "2024-03-22T12:00:00Z",
                "completed_at": "2024-03-22T12:05:30Z",
                "sections": [
                    {
                        "title": "Focus Time Analysis",
                        "content": "Your focus time has increased by 15% this week..."
                    },
                    {
                        "title": "Task Completion Metrics",
                        "content": "You completed 28 tasks this week with an average completion time of..."
                    }
                ]
            }
        }
