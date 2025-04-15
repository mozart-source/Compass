from fastapi import APIRouter, Depends, HTTPException, Query, WebSocket, WebSocketDisconnect
from typing import List, Optional, Dict
from app.schemas.system_metric_schemas import SystemMetricCreate, SystemMetricResponse, SystemMetricListResponse
from data_layer.repos.system_metric_repo import SystemMetricRepository
from data_layer.models.system_metric_model import SystemMetric
from utils.jwt import extract_user_id_from_token
from datetime import datetime
import json
from data_layer.cache.pubsub_manager import pubsub_manager
from fastapi.encoders import jsonable_encoder

router = APIRouter(prefix="/system-metrics", tags=["System Metrics"])
metric_repo = SystemMetricRepository()


@router.post("/", response_model=SystemMetricResponse)
def create_metric(metric: SystemMetricCreate, user_id: str = Depends(extract_user_id_from_token)):
    metric_data = metric.dict()
    metric_data["user_id"] = user_id
    if not metric_data.get("timestamp"):
        metric_data["timestamp"] = datetime.utcnow()
    metric_obj = SystemMetric(**metric_data)
    metric_id = metric_repo.create_metric(metric_obj)
    # Publish to Redis pub/sub for real-time update
    import asyncio
    serializable_data = jsonable_encoder(metric_obj.model_dump())
    asyncio.create_task(pubsub_manager.publish(
        user_id, "system_metric", serializable_data))
    return SystemMetricResponse(id=str(metric_id), user_id=user_id, **metric.dict())


@router.get("/", response_model=SystemMetricListResponse)
def list_metrics(user_id: str = Depends(extract_user_id_from_token)):
    metrics = metric_repo.find_by_user(user_id)
    return SystemMetricListResponse(metrics=[SystemMetricResponse(id=str(m.id), user_id=m.user_id, metric_type=m.metric_type, value=m.value, timestamp=m.timestamp, metadata=m.metadata) for m in metrics])


@router.get("/range", response_model=SystemMetricListResponse)
def metrics_by_type_and_range(
    metric_type: str = Query(...),
    start: datetime = Query(...),
    end: datetime = Query(...),
    user_id: str = Depends(extract_user_id_from_token)
):
    metrics = metric_repo.find_by_type_and_range(
        user_id, metric_type, start, end)
    return SystemMetricListResponse(metrics=[SystemMetricResponse(id=str(m.id), user_id=m.user_id, metric_type=m.metric_type, value=m.value, timestamp=m.timestamp, metadata=m.metadata) for m in metrics])


@router.websocket("/ws")
async def system_metrics_ws(websocket: WebSocket):
    await websocket.accept()
    user_id = None
    try:
        token = websocket.headers.get("authorization")
        if not token:
            await websocket.close(code=4001)
            return
        user_id = extract_user_id_from_token(token)
        # Subscribe to Redis pub/sub for this user

        async def ws_callback(message):
            await websocket.send_text(message)
        await pubsub_manager.subscribe(user_id, ws_callback)
        while True:
            data = await websocket.receive_text()
            metric_data = json.loads(data)
            metric_data["user_id"] = user_id
            if not metric_data.get("timestamp"):
                metric_data["timestamp"] = datetime.utcnow().isoformat()
            metric = SystemMetric(**metric_data)
            metric_repo.create_metric(metric)
            # Publish to Redis pub/sub for real-time update
            serializable_metric_data = jsonable_encoder(metric_data)
            await pubsub_manager.publish(user_id, "system_metric", serializable_metric_data)
    except WebSocketDisconnect:
        if user_id:
            await pubsub_manager.unsubscribe(user_id)
    except Exception:
        if websocket.client_state.value == 1:  # OPEN
            await websocket.close(code=4002)
        if user_id:
            await pubsub_manager.unsubscribe(user_id)


@router.get("/summary", response_model=List[Dict])
def summary_metrics(
    period: str = Query(
        "daily", description="Aggregation period: daily, weekly, monthly"),
    metric_type: Optional[str] = Query(
        None, description="Metric type to filter (optional)"),
    start: Optional[datetime] = Query(
        None, description="Start datetime (optional)"),
    end: Optional[datetime] = Query(
        None, description="End datetime (optional)"),
    user_id: str = Depends(extract_user_id_from_token)
):
    """
    Get aggregated system metrics for the dashboard. Returns sum, avg, min, max, count per period and metric_type.
    """
    results = metric_repo.aggregate_metrics(
        user_id, period=period, metric_type=metric_type, start=start, end=end)
    return results
