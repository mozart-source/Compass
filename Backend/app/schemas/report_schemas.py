from pydantic import BaseModel, Field
from typing import Dict, Any, Optional, List
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


class ReportCreate(BaseModel):
    """Schema for creating a new report."""
    title: str = Field(..., description="Report title")
    type: ReportType = Field(..., description="Report type")
    parameters: Optional[Dict[str, Any]] = Field(
        None, description="Additional parameters for report generation")
    time_range: Optional[Dict[str, str]] = Field(
        None, description="Time range for report data (start_date, end_date)")
    custom_prompt: Optional[str] = Field(
        None, description="Custom prompt for report generation")


class ReportSection(BaseModel):
    """Schema for a report section."""
    title: str = Field(..., description="Section title")
    content: str = Field(..., description="Section content")
    type: str = Field(
        "text", description="Section content type (text, chart, table, etc.)")
    data: Optional[Dict[str, Any]] = Field(
        None, description="Additional data for the section")


class ReportUpdate(BaseModel):
    """Schema for updating a report."""
    title: Optional[str] = Field(None, description="Report title")
    parameters: Optional[Dict[str, Any]] = Field(
        None, description="Additional parameters for report generation")
    time_range: Optional[Dict[str, str]] = Field(
        None, description="Time range for report data (start_date, end_date)")
    custom_prompt: Optional[str] = Field(
        None, description="Custom prompt for report generation")


class ReportResponse(BaseModel):
    """Schema for report response."""
    id: str = Field(..., description="Report ID")
    title: str = Field(..., description="Report title")
    type: ReportType = Field(..., description="Report type")
    status: ReportStatus = Field(..., description="Report status")
    user_id: str = Field(..., description="User ID")
    created_at: datetime = Field(..., description="Creation timestamp")
    updated_at: datetime = Field(..., description="Last update timestamp")
    completed_at: Optional[datetime] = Field(
        None, description="Completion timestamp")
    parameters: Optional[Dict[str, Any]] = Field(
        None, description="Report parameters")
    time_range: Optional[Dict[str, str]] = Field(
        None, description="Time range for report data")
    custom_prompt: Optional[str] = Field(
        None, description="Custom prompt used for generation")
    summary: Optional[str] = Field(None, description="Report summary")
    content: Optional[Dict[str, Any]] = Field(
        None, description="Report content")
    sections: Optional[List[ReportSection]] = Field(
        None, description="Report sections")
    error: Optional[str] = Field(
        None, description="Error message if generation failed")

    class Config:
        """Pydantic config."""
        from_attributes = True
        use_enum_values = True


class ReportProgressUpdate(BaseModel):
    """Schema for report generation progress updates via WebSocket."""
    report_id: str = Field(..., description="Report ID")
    progress: float = Field(..., description="Progress percentage (0.0-1.0)")
    status: str = Field(...,
                        description="Current status (generating, completed, failed)")
    message: str = Field(..., description="Progress message")


class ReportListResponse(BaseModel):
    """Schema for paginated report list response."""
    reports: List[ReportResponse] = Field(..., description="List of reports")
    total: int = Field(...,
                       description="Total number of reports matching filters")
    page: int = Field(..., description="Current page number")
    limit: int = Field(..., description="Number of reports per page")


class ReportTypeInfo(BaseModel):
    """Schema for report type information."""
    type: str = Field(..., description="Report type identifier")
    name: str = Field(..., description="Report type display name")
    description: str = Field(..., description="Report type description")


class ReportGenerateRequest(BaseModel):
    report_id: str
