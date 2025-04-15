from typing import Optional, Dict, ClassVar
from datetime import datetime
from pydantic import Field
from .base_model import MongoBaseModel


class SystemMetric(MongoBaseModel):
    user_id: str = Field(..., description="User ID")
    metric_type: str = Field(
        ..., description="Type of system metric (keyboard, screen_time, app_usage, productivity_score, etc.)")
    value: float = Field(...,
                         description="Metric value (float/int/str as float)")
    timestamp: datetime = Field(
        default_factory=datetime.utcnow, description="Timestamp of the metric")
    metadata: Optional[Dict] = Field(
        default=None, description="Optional extra info")
    
    collection_name: ClassVar[str] = "system_metrics"
