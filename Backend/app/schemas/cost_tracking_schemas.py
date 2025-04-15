from typing import Dict, Any, Optional, List
from pydantic import BaseModel, Field
from datetime import datetime


class CostTrackingEntryCreate(BaseModel):
    model_id: str
    input_tokens: int = 0
    output_tokens: int = 0
    input_cost: float = 0.0
    output_cost: float = 0.0
    total_cost: float = 0.0
    success: bool = True
    request_id: str
    timestamp: Optional[datetime] = None
    metadata: Optional[Dict[str, Any]] = Field(default_factory=dict)


class CostTrackingEntryResponse(CostTrackingEntryCreate):
    id: str
    user_id: str


class CostSummaryResponse(BaseModel):
    total_cost: float
    total_tokens: int
    successful_requests: int
    failed_requests: int
    models_used: List[str] = []


class CostTrendResponse(BaseModel):
    interval: str
    total_cost: float
    total_tokens: int
    successful_requests: int
    failed_requests: int
