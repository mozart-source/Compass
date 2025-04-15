package dto

import (
	"time"

	"github.com/google/uuid"
)

type TimelineItem struct {
	ID          uuid.UUID  `json:"id"`
	Title       string     `json:"title"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Type        string     `json:"type"` // "habit", "task", "todo", "event"
	IsCompleted bool       `json:"is_completed"`
}

type DashboardMetricsResponse struct {
	Habits        HabitsDashboardMetrics   `json:"habits"`
	Tasks         TasksDashboardMetrics    `json:"tasks"`
	Todos         TodosDashboardMetrics    `json:"todos"`
	Calendar      CalendarDashboardMetrics `json:"calendar"`
	User          UserDashboardMetrics     `json:"user"`
	DailyTimeline []TimelineItem           `json:"daily_timeline"`
	HabitHeatmap  map[string]int           `json:"habit_heatmap"`
	Timestamp     time.Time                `json:"timestamp"`
}

type HabitsDashboardMetrics struct {
	Total     int `json:"total"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Streak    int `json:"streak"`
}

type TasksDashboardMetrics struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Overdue   int `json:"overdue"`
}

type TodosDashboardMetrics struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Overdue   int `json:"overdue"`
}

type CalendarDashboardMetrics struct {
	Upcoming int `json:"upcoming"`
	Total    int `json:"total"`
}

type UserDashboardMetrics struct {
	ActivitySummary map[string]int `json:"activity_summary"`
}
