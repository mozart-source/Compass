package habits

import (
	"time"

	"github.com/google/uuid"
)

// HabitAnalytics represents an analytics record for habit-related activities
type HabitAnalytics struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	HabitID   uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Action    string    `gorm:"type:varchar(50);not null"`
	Timestamp time.Time `gorm:"not null;default:now()"`
	Metadata  string    `gorm:"type:jsonb"`
}

// TableName specifies the table name for the HabitAnalytics model
func (HabitAnalytics) TableName() string {
	return "habit_analytics"
}

// AnalyticsFilter defines filtering options for habit analytics
type AnalyticsFilter struct {
	HabitID   *uuid.UUID
	UserID    *uuid.UUID
	Action    *string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int
}

// Analytics types for service
type RecordHabitActivityInput struct {
	HabitID   uuid.UUID              `json:"habit_id"`
	UserID    uuid.UUID              `json:"user_id"`
	Action    string                 `json:"action"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
}

type HabitActivitySummary struct {
	HabitID      uuid.UUID      `json:"habit_id"`
	ActionCounts map[string]int `json:"action_counts"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	TotalActions int            `json:"total_actions"`
}

type UserHabitActivitySummary struct {
	UserID       uuid.UUID      `json:"user_id"`
	ActionCounts map[string]int `json:"action_counts"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	TotalActions int            `json:"total_actions"`
}

// Common analytics actions
const (
	ActionHabitCreated        = "habit_created"
	ActionHabitUpdated        = "habit_updated"
	ActionHabitDeleted        = "habit_deleted"
	ActionHabitCompleted      = "habit_completed"
	ActionHabitUncompleted    = "habit_uncompleted"
	ActionStreakStarted       = "streak_started"
	ActionStreakBroken        = "streak_broken"
	ActionStreakMilestone     = "streak_milestone"
	ActionHabitReminderSent   = "habit_reminder_sent"
	ActionHabitReminderOpened = "habit_reminder_opened"
	ActionHabitView           = "habit_view"
	ActionHabitListView       = "habit_list_view"
	ActionHabitStats          = "habit_stats_view"
	ActionHabitHeatmapView    = "habit_heatmap_view"
	ActionStreakHistoryView   = "streak_history_view"
	ActionHabitDueTodayView   = "habits_due_today_view"
)
