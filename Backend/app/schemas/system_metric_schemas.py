from pydantic import BaseModel, Field
from typing import Optional, Dict, List
from datetime import datetime


class SystemMetricBase(BaseModel):
    metric_type: str = Field(...)
    value: float = Field(...)
    timestamp: Optional[datetime] = None
    metadata: Optional[Dict] = None


class SystemMetricCreate(SystemMetricBase):
    pass


class SystemMetricResponse(SystemMetricBase):
    id: str
    user_id: str

    class Config:
        from_attributes = True


class SystemMetricListResponse(BaseModel):
    metrics: List[SystemMetricResponse]


class SystemMetricFilter(BaseModel):
    user_id: str
    metric_type: Optional[str] = None
    start_date: Optional[datetime] = None
    end_date: Optional[datetime] = None
