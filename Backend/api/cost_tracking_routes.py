from fastapi import APIRouter, Depends, HTTPException, Query
from typing import List, Optional
from datetime import datetime
from app.schemas.cost_tracking_schemas import CostTrackingEntryCreate, CostTrackingEntryResponse, CostSummaryResponse, CostTrendResponse
from data_layer.repos.cost_tracking_repo import CostTrackingRepository
from utils.jwt import extract_user_id_from_token
from fastapi.encoders import jsonable_encoder

router = APIRouter(prefix="/cost-tracking", tags=["Cost Tracking"])
repo = CostTrackingRepository()


@router.post("/", response_model=CostTrackingEntryResponse)
async def create_cost_entry(
    data: CostTrackingEntryCreate,
    user_id: str = Depends(extract_user_id_from_token)
):
    entry_data = data.model_dump()
    entry_data["user_id"] = user_id
    if not entry_data.get("timestamp"):
        entry_data["timestamp"] = datetime.utcnow()
    entry_id = await repo.create_tracking_entry(entry_data)
    entry = entry_data.copy()
    entry["id"] = str(entry_id)
    return CostTrackingEntryResponse(**jsonable_encoder(entry))


@router.get("/summary", response_model=CostSummaryResponse)
async def get_cost_summary(
    start: Optional[datetime] = Query(None),
    end: Optional[datetime] = Query(None),
    user_id: str = Depends(extract_user_id_from_token)
):
    summary = await repo.get_user_cost_summary(user_id, start or datetime.utcnow().replace(day=1), end or datetime.utcnow())
    return CostSummaryResponse(**jsonable_encoder(summary))


@router.get("/trends", response_model=List[CostTrendResponse])
async def get_cost_trends(
    interval: str = Query(
        "daily", description="Interval: hourly, daily, weekly, monthly"),
    start: Optional[datetime] = Query(None),
    end: Optional[datetime] = Query(None),
    user_id: str = Depends(extract_user_id_from_token)
):
    trends = await repo.get_cost_trends(user_id=user_id, interval=interval, start_time=start, end_time=end)
    return [CostTrendResponse(**jsonable_encoder(t)) for t in trends]


@router.get("/", response_model=List[CostTrackingEntryResponse])
async def list_cost_entries(user_id: str = Depends(extract_user_id_from_token)):
    entries = await repo.async_find_many({"user_id": user_id})
    return [CostTrackingEntryResponse(**jsonable_encoder(e.model_dump(), exclude_unset=True)) for e in entries]
