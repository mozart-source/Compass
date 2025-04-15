from typing import List, Optional, Dict, Any, ClassVar
from pydantic import Field, field_validator
from data_layer.models.base_model import MongoBaseModel
from datetime import datetime
from enum import Enum


class ModelType(str, Enum):
    """Types of AI models."""
    TEXT_GENERATION = "text-generation"
    EMBEDDING = "embedding"
    CLASSIFICATION = "classification"
    SUMMARIZATION = "summarization"
    RAG = "rag"
    VISION = "vision"
    AUDIO = "audio"
    MULTIMODAL = "multimodal"


class ModelProvider(str, Enum):
    """AI model providers."""
    OPENAI = "openai"
    ANTHROPIC = "anthropic"
    GOOGLE = "google"
    HUGGINGFACE = "huggingface"
    CUSTOM = "custom"
    AZURE = "azure"
    AWS = "aws"
    COHERE = "cohere"


class BillingType(str, Enum):
    """Billing types for AI models."""
    PAY_AS_YOU_GO = "pay-as-you-go"
    PROVISIONED = "provisioned"
    QUOTA_BASED = "quota-based"
    HYBRID = "hybrid"


class AIModel(MongoBaseModel):
    """AI model metadata and configuration."""

    name: str = Field(..., description="Model name")
    version: str = Field(..., description="Model version")
    type: ModelType = Field(..., description="Type of model")
    provider: ModelProvider = Field(..., description="Model provider")
    status: str = Field(
        "active", description="Model status (active, inactive, deprecated)")
    capabilities: Dict[str, Any] = Field(
        default_factory=dict, description="Model capabilities")
    config: Dict[str, Any] = Field(
        default_factory=dict, description="Model configuration")
    metrics: Dict[str, Any] = Field(
        default_factory=dict, description="Model performance metrics")

    # Billing configuration
    billing_type: BillingType = Field(
        default=BillingType.PAY_AS_YOU_GO, description="Billing type for this model")
    input_token_cost_per_million: float = Field(
        default=0.0, description="Cost per million input tokens")
    output_token_cost_per_million: float = Field(
        default=0.0, description="Cost per million output tokens")
    provisioned_capacity: Optional[int] = Field(
        None, description="Provisioned capacity in tokens per hour")
    provisioned_cost_per_hour: Optional[float] = Field(
        None, description="Cost per hour for provisioned capacity")
    quota_limit: Optional[int] = Field(
        None, description="Token quota limit for quota-based billing")
    quota_reset_interval: Optional[str] = Field(
        None, description="Interval for quota reset (daily, weekly, monthly)")

    # Set collection name
    collection_name: ClassVar[str] = "ai_models"


class ModelUsage(MongoBaseModel):
    """Model usage statistics for tracking and billing."""

    model_id: str = Field(..., description="ID of the AI model used")
    model_name: str = Field(..., description="Name of the model")
    user_id: Optional[str] = Field(None, description="User who used the model")
    session_id: Optional[str] = Field(None, description="Session ID")
    request_type: str = Field(...,
                              description="Type of request (generation, classification, etc.)")
    tokens_in: int = Field(0, description="Number of input tokens")
    tokens_out: int = Field(0, description="Number of output tokens")
    latency_ms: int = Field(0, description="Request latency in milliseconds")
    success: bool = Field(
        True, description="Whether the request was successful")
    error: Optional[str] = Field(
        None, description="Error message if request failed")

    # Cost tracking
    input_cost: float = Field(0.0, description="Cost for input tokens")
    output_cost: float = Field(0.0, description="Cost for output tokens")
    total_cost: float = Field(0.0, description="Total cost for this usage")
    billing_type: BillingType = Field(
        default=BillingType.PAY_AS_YOU_GO, description="Billing type used")
    quota_applied: bool = Field(
        False, description="Whether quota was applied")
    quota_exceeded: bool = Field(
        False, description="Whether quota was exceeded")

    # Request metadata
    request_id: str = Field(..., description="Unique request identifier")
    endpoint: str = Field(..., description="API endpoint used")
    client_ip: Optional[str] = Field(None, description="Client IP address")
    user_agent: Optional[str] = Field(None, description="User agent string")
    organization_id: Optional[str] = Field(None, description="Organization ID")

    # Set collection name
    collection_name: ClassVar[str] = "model_usage"
