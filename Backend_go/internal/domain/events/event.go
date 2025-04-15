package events

import (
	"time"

	"github.com/google/uuid"
)

// Dashboard event types
const (
	EventTypeUserActivity    = "user_activity"
	EventTypeHabitUpdate     = "habit_update"
	EventTypeTodoUpdate      = "todo_update"
	EventTypeCalendarUpdate  = "calendar_update"
	EventTypeDashboardUpdate = "dashboard_update"
)

// DashboardEvent represents a dashboard-related event
type DashboardEvent struct {
	EventType string      `json:"event_type"`
	UserID    uuid.UUID   `json:"user_id"`
	EntityID  uuid.UUID   `json:"entity_id"`
	Timestamp time.Time   `json:"timestamp"`
	Details   interface{} `json:"details,omitempty"`
}

// DashboardEventTypes defines standard event types for dashboard events
const (
	DashboardEventMetricsUpdate   = "metrics_update"
	DashboardEventCacheInvalidate = "cache_invalidate"
)

// DashboardMetrics represents the complete dashboard metrics
type DashboardMetrics struct {
	User      interface{} `json:"user"`
	Habits    interface{} `json:"habits"`
	Todos     interface{} `json:"todos"`
	Calendar  interface{} `json:"calendar"`
	Timestamp time.Time   `json:"timestamp"`
}
