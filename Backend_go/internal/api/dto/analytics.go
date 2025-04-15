package dto

import (
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/user"
	"github.com/google/uuid"
)

// UserAnalyticsFilter represents the filter parameters for analytics queries
type UserAnalyticsFilter struct {
	StartTime string `form:"start_time" json:"start_time" binding:"required"`
	EndTime   string `form:"end_time" json:"end_time" binding:"required"`
	Page      int    `form:"page" json:"page" binding:"min=0"`
	PageSize  int    `form:"page_size" json:"page_size" binding:"min=1,max=100"`
	Action    string `form:"action" json:"action,omitempty"`
}

// RecordUserActivityRequest represents the request to record user activity
type RecordUserActivityRequest struct {
	Action   string                 `json:"action" binding:"required"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RecordSessionActivityRequest represents the request to record session activity
type RecordSessionActivityRequest struct {
	Action     string                 `json:"action" binding:"required"`
	DeviceInfo string                 `json:"device_info,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// UserAnalyticsResponse represents a single user analytics entry
type UserAnalyticsResponse struct {
	ID        uuid.UUID              `json:"id"`
	UserID    uuid.UUID              `json:"user_id"`
	Action    string                 `json:"action"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SessionAnalyticsResponse represents a single session analytics entry
type SessionAnalyticsResponse struct {
	ID         uuid.UUID              `json:"id"`
	SessionID  string                 `json:"session_id"`
	UserID     uuid.UUID              `json:"user_id"`
	Action     string                 `json:"action"`
	DeviceInfo string                 `json:"device_info,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// UserAnalyticsListResponse represents the paginated response for user analytics
type UserAnalyticsListResponse struct {
	Analytics  []UserAnalyticsResponse `json:"analytics"`
	TotalCount int64                   `json:"total_count"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
}

// SessionAnalyticsListResponse represents the paginated response for session analytics
type SessionAnalyticsListResponse struct {
	Analytics  []SessionAnalyticsResponse `json:"analytics"`
	TotalCount int64                      `json:"total_count"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
}

// UserActivitySummaryResponse represents a summary of user activity
type UserActivitySummaryResponse struct {
	UserID       uuid.UUID      `json:"user_id"`
	ActionCounts map[string]int `json:"action_counts"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	TotalActions int            `json:"total_actions"`
}

// TaskAnalyticsResponse represents a single task analytics entry
type TaskAnalyticsResponse struct {
	ID        uuid.UUID              `json:"id"`
	TaskID    uuid.UUID              `json:"task_id"`
	UserID    uuid.UUID              `json:"user_id"`
	Action    string                 `json:"action"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TaskAnalyticsListResponse represents the paginated response for task analytics
type TaskAnalyticsListResponse struct {
	Analytics  []TaskAnalyticsResponse `json:"analytics"`
	TotalCount int64                   `json:"total_count"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
}

// EventAnalyticsResponse represents a single event analytics entry
type EventAnalyticsResponse struct {
	ID        uuid.UUID              `json:"id"`
	EventID   uuid.UUID              `json:"event_id"`
	UserID    uuid.UUID              `json:"user_id"`
	Action    string                 `json:"action"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// EventAnalyticsListResponse represents the paginated response for event analytics
type EventAnalyticsListResponse struct {
	Analytics  []EventAnalyticsResponse `json:"analytics"`
	TotalCount int64                    `json:"total_count"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
}

// UserActivityInput represents the input for recording user activity
type UserActivityInput struct {
	UserID    uuid.UUID              `json:"user_id"`
	Action    string                 `json:"action"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
}

// ToDomain converts UserActivityInput to user.RecordUserActivityInput
func (i UserActivityInput) ToDomain() user.RecordUserActivityInput {
	return user.RecordUserActivityInput{
		UserID:    i.UserID,
		Action:    i.Action,
		Metadata:  i.Metadata,
		Timestamp: i.Timestamp,
	}
}
