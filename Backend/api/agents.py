from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.ext.asyncio import AsyncSession
from Backend.data_layer.database.connection import get_db
from Backend.data_layer.repositories.agent_repository import AgentRepository
from Backend.app.schemas.agent_schemas import (
    AgentActionCreate, AgentActionResponse, AgentFeedbackCreate,
    AgentFeedbackResponse, AgentInteractionCreate, AgentInteractionResponse
)
from Backend.api.auth import get_current_user
from typing import Dict, List, Optional
from datetime import datetime

router = APIRouter(prefix="/agents", tags=["agents"])

@router.post("/actions", response_model=AgentActionResponse)
async def create_agent_action(
    action: AgentActionCreate,
    db: AsyncSession = Depends(get_db),
    current_user = Depends(get_current_user)
):
    """Create a new agent action."""
    repo = AgentRepository(db)
    result = await repo.create_agent_action(**action.dict())
    return result

@router.post("/feedback", response_model=AgentFeedbackResponse)
async def create_agent_feedback(
    feedback: AgentFeedbackCreate,
    db: AsyncSession = Depends(get_db),
    current_user = Depends(get_current_user)
):
    """Create feedback for an agent action."""
    repo = AgentRepository(db)
    result = await repo.create_agent_feedback(**feedback.dict())
    return result

@router.post("/interactions", response_model=AgentInteractionResponse)
async def create_agent_interaction(
    interaction: AgentInteractionCreate,
    db: AsyncSession = Depends(get_db),
    current_user = Depends(get_current_user)
):
    """Create a new agent interaction."""
    repo = AgentRepository(db)
    result = await repo.create_agent_interaction(**interaction.dict())
    return result

@router.get("/interactions", response_model=List[AgentInteractionResponse])
async def get_agent_interactions(
    agent_type: Optional[str] = None,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    limit: int = 100,
    db: AsyncSession = Depends(get_db),
    current_user = Depends(get_current_user)
):
    """Get agent interactions with optional filtering."""
    repo = AgentRepository(db)
    interactions = await repo.get_agent_interactions(
        user_id=current_user.id,
        agent_type=agent_type,
        start_date=start_date,
        end_date=end_date,
        limit=limit
    )
    return interactions

@router.get("/metrics")
async def get_agent_performance_metrics(
    agent_type: Optional[str] = None,
    time_period: Optional[int] = None,
    db: AsyncSession = Depends(get_db),
    current_user = Depends(get_current_user)
):
    """Get performance metrics for agents."""
    repo = AgentRepository(db)
    metrics = await repo.get_agent_performance_metrics(
        agent_type=agent_type,
        time_period=time_period
    )
    return metrics