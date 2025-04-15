from pydantic import BaseModel
from typing import Optional, Dict, Any, List
from datetime import datetime


class TimelineItem(BaseModel):
    id: str
    title: str
    start_time: datetime
    end_time: Optional[datetime] = None
    type: str  # "habit", "task", "todo", "event"
    is_completed: bool = False


class DailyFocusItem(BaseModel):
    day: str
    minutes: int


class FocusMetrics(BaseModel):
    total_focus_seconds: int = 0
    streak: int = 0
    longest_streak: int = 0
    sessions: int = 0
    daily_breakdown: List[DailyFocusItem] = []


class DashboardMetrics(BaseModel):
    habits: Optional[Dict[str, Any]] = None
    tasks: Optional[Dict[str, Any]] = None
    todos: Optional[Dict[str, Any]] = None
    calendar: Optional[Dict[str, Any]] = None
    focus: Optional[FocusMetrics] = None
    mood: Optional[Dict[str, Any]] = None
    ai_usage: Optional[Dict[str, Any]] = None
    system_metrics: Optional[Dict[str, Any]] = None
    goals: Optional[Dict[str, Any]] = None
    projects: Optional[Dict[str, Any]] = None
    workflows: Optional[Dict[str, Any]] = None
    user: Optional[Dict[str, Any]] = None
    notes: Optional[Dict[str, Any]] = None
    journals: Optional[Dict[str, Any]] = None
    cost: Optional[Dict[str, Any]] = None
    daily_timeline: Optional[List[TimelineItem]] = None
    habit_heatmap: Optional[Dict[str, int]] = None
    timestamp: Optional[datetime] = None
