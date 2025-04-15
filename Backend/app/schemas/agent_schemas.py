from pydantic import BaseModel, Field
from typing import Dict, List, Optional
from datetime import datetime
from enum import Enum


class AgentType(str, Enum):
    TASK = "task"
    WORKFLOW = "workflow"
    RESOURCE = "resource"
    COLLABORATION = "collaboration"
    PRODUCTIVITY = "productivity"
    ANALYSIS = "analysis"

class AgentActionBase(BaseModel):
    agent_type: str
    user_id: Optional[int] = None
    request_id: Optional[str] = None
    action_data: Dict = Field(default_factory=dict)
    status: Optional[str] = None
    error_message: Optional[str] = None

class AgentActionCreate(AgentActionBase):
    pass

class AgentActionResponse(AgentActionBase):
    id: int
    result: Optional[Dict] = None
    timestamp: datetime

    class Config:
        from_attributes = True

class AgentFeedbackCreate(BaseModel):
    agent_action_id: int
    user_id: int
    feedback_score: int
    feedback_text: Optional[str] = None

class AgentFeedbackResponse(AgentFeedbackCreate):
    id: int
    timestamp: datetime

    class Config:
        from_attributes = True

class AgentInteractionBase(BaseModel):
    workflow_id: int
    agent_id: int
    interaction_type: str
    confidence_score: Optional[float] = None
    input_data: Optional[Dict] = Field(default_factory=dict)
    output_data: Optional[Dict] = Field(default_factory=dict)
    execution_time: Optional[float] = None
    performance_metrics: Optional[Dict] = Field(default_factory=dict)
    optimization_suggestions: Optional[Dict] = Field(default_factory=dict)

class AgentInteractionCreate(AgentInteractionBase):
    pass

class AgentInteractionResponse(AgentInteractionBase):
    id: int
    status: str
    error_message: Optional[str] = None
    created_at: datetime

    class Config:
        from_attributes = True