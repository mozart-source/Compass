from fastapi import APIRouter, Depends, HTTPException
from typing import List
from app.schemas.goal_schemas import GoalCreate, GoalUpdate, GoalResponse, GoalListResponse
from data_layer.repos.goal_repo import GoalRepository
from data_layer.models.goal_model import Goal
from utils.jwt import extract_user_id_from_token

router = APIRouter(prefix="/goals", tags=["Goals"])
goal_repo = GoalRepository()


@router.get("/", response_model=GoalListResponse)
def list_goals(user_id: str = Depends(extract_user_id_from_token)):
    goals = goal_repo.find_by_user(user_id)
    return GoalListResponse(goals=[GoalResponse(**g.dict()) for g in goals])


@router.post("/", response_model=GoalResponse)
def create_goal(goal: GoalCreate, user_id: str = Depends(extract_user_id_from_token)):
    goal_data = goal.dict()
    goal_data["user_id"] = user_id
    new_goal = goal_repo.create_goal(Goal(**goal_data))
    created = goal_repo.find_by_id(new_goal)
    if not created:
        raise HTTPException(status_code=500, detail="Goal creation failed")
    return GoalResponse(**created.dict())


@router.put("/{goal_id}", response_model=GoalResponse)
def update_goal(goal_id: str, goal: GoalUpdate, user_id: str = Depends(extract_user_id_from_token)):
    existing = goal_repo.find_by_id(goal_id)
    if not existing or existing.user_id != user_id:
        raise HTTPException(status_code=404, detail="Goal not found")
    updated = goal_repo.update_goal(goal_id, goal.dict(exclude_unset=True))
    if not updated:
        raise HTTPException(status_code=500, detail="Goal update failed")
    return GoalResponse(**updated.dict())


@router.delete("/{goal_id}")
def delete_goal(goal_id: str, user_id: str = Depends(extract_user_id_from_token)):
    existing = goal_repo.find_by_id(goal_id)
    if not existing or existing.user_id != user_id:
        raise HTTPException(status_code=404, detail="Goal not found")
    success = goal_repo.delete_goal(goal_id)
    if not success:
        raise HTTPException(status_code=500, detail="Goal deletion failed")
    return {"success": True}
