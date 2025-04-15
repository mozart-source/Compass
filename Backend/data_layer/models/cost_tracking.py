from typing import Dict, Any, Optional, ClassVar
from pydantic import Field
from data_layer.models.base_model import MongoBaseModel
from datetime import datetime


class CostTrackingEntry(MongoBaseModel):
    """Model for tracking AI model usage costs."""

    model_id: str = Field(..., description="ID of the AI model used")
    user_id: str = Field(...,
                         description="ID of the user who made the request")
    input_tokens: int = Field(0, description="Number of input tokens")
    output_tokens: int = Field(0, description="Number of output tokens")
    input_cost: float = Field(0.0, description="Cost for input tokens")
    output_cost: float = Field(0.0, description="Cost for output tokens")
    total_cost: float = Field(0.0, description="Total cost for this usage")
    success: bool = Field(
        True, description="Whether the request was successful")
    request_id: str = Field(..., description="Unique request identifier")
    timestamp: datetime = Field(
        default_factory=datetime.utcnow, description="When the request was made")
    metadata: Dict[str, Any] = Field(
        default_factory=dict,
        description="Additional metadata about the request"
    )

    # Set collection name
    collection_name: ClassVar[str] = "cost_tracking"

    class Config:
        """Pydantic model configuration."""
        json_encoders = {
            datetime: lambda dt: dt.isoformat()
        }
        json_schema_extra = {
            "example": {
                "model_id": "model123",
                "user_id": "user123",
                "input_tokens": 100,
                "output_tokens": 50,
                "input_cost": 0.0025,
                "output_cost": 0.00375,
                "total_cost": 0.00625,
                "success": True,
                "request_id": "req123",
                "timestamp": "2024-03-15T12:00:00Z",
                "metadata": {
                    "endpoint": "chat/completions",
                    "client_ip": "127.0.0.1",
                    "user_agent": "Python/3.9",
                    "organization_id": "org123"
                }
            }
        }
