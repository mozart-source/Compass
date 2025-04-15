package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateHabitRequest represents the request to create a new habit
type CreateHabitRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	StartDay    time.Time  `json:"start_day" binding:"required"`
	EndDay      *time.Time `json:"end_day"`
}

// UpdateHabitRequest represents the request to update an existing habit
type UpdateHabitRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	StartDay    *time.Time `json:"start_day,omitempty"`
	EndDay      *time.Time `json:"end_day,omitempty"`
}

// HabitCompletionRequest represents the request to mark a habit as completed
type HabitCompletionRequest struct {
	CompletionDate *time.Time `json:"completion_date,omitempty"`
}

// HabitResponse represents a habit in API responses
type HabitResponse struct {
	ID                uuid.UUID  `json:"id"`
	UserID            uuid.UUID  `json:"user_id"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	StartDay          time.Time  `json:"start_day"`
	EndDay            *time.Time `json:"end_day,omitempty"`
	CurrentStreak     int        `json:"current_streak"`
	StreakStartDate   *time.Time `json:"streak_start_date,omitempty"`
	LongestStreak     int        `json:"longest_streak"`
	IsCompleted       bool       `json:"is_completed"`
	LastCompletedDate *time.Time `json:"last_completed_date,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	StreakQuality     float64    `json:"streak_quality"`
}

// HabitListResponse represents the response for listing habits
type HabitListResponse struct {
	Habits     []HabitResponse `json:"habits"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
}

// StreakHistoryResponse represents a streak history record in API responses
type StreakHistoryResponse struct {
	ID            uuid.UUID `json:"id"`
	HabitID       uuid.UUID `json:"habit_id"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	StreakLength  int       `json:"streak_length"`
	CompletedDays int       `json:"completed_days"`
	CreatedAt     time.Time `json:"created_at"`
}

// HabitStatsResponse represents statistics about habits
type HabitStatsResponse struct {
	TotalHabits     int `json:"total_habits"`
	ActiveHabits    int `json:"active_habits"`
	CompletedHabits int `json:"completed_habits"`
}

// HeatmapResponse represents habit completion heatmap data
type HeatmapResponse struct {
	Data     map[string]int `json:"data"`
	Period   string         `json:"period"`
	MinValue int            `json:"min_value"`
	MaxValue int            `json:"max_value"`
}

// HabitAnalyticsFilter represents the filter parameters for habit analytics queries
type HabitAnalyticsFilter struct {
	StartTime string `form:"start_time" json:"start_time" binding:"required"`
	EndTime   string `form:"end_time" json:"end_time" binding:"required"`
	Page      int    `form:"page" json:"page" binding:"min=0"`
	PageSize  int    `form:"page_size" json:"page_size" binding:"min=1,max=100"`
	Action    string `form:"action" json:"action,omitempty"`
}

// RecordHabitActivityRequest represents the request to record habit activity
type RecordHabitActivityRequest struct {
	Action   string                 `json:"action" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HabitAnalyticsResponse represents a single habit analytics entry
type HabitAnalyticsResponse struct {
	ID        uuid.UUID              `json:"id"`
	HabitID   uuid.UUID              `json:"habit_id"`
	UserID    uuid.UUID              `json:"user_id"`
	Action    string                 `json:"action"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// HabitAnalyticsListResponse represents the paginated response for habit analytics
type HabitAnalyticsListResponse struct {
	Analytics  []HabitAnalyticsResponse `json:"analytics"`
	TotalCount int64                    `json:"total_count"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
}

// HabitActivitySummaryResponse represents a summary of habit activity
type HabitActivitySummaryResponse struct {
	HabitID      uuid.UUID      `json:"habit_id"`
	ActionCounts map[string]int `json:"action_counts"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	TotalActions int            `json:"total_actions"`
}
