package task

import (
	"time"

	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskStatus string

const (
	TaskStatusUpcoming    TaskStatus = "Upcoming"
	TaskStatusInProgress  TaskStatus = "In Progress"
	TaskStatusCompleted   TaskStatus = "Completed"
	TaskStatusCancelled   TaskStatus = "Cancelled"
	TaskStatusBlocked     TaskStatus = "Blocked"
	TaskStatusUnderReview TaskStatus = "Under Review"
	TaskStatusDeferred    TaskStatus = "Deferred"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "Low"
	TaskPriorityMedium TaskPriority = "Medium"
	TaskPriorityHigh   TaskPriority = "High"
	TaskPriorityUrgent TaskPriority = "Urgent"
)

type UUIDSlice []uuid.UUID

// Value implements the driver.Valuer interface for UUIDSlice
func (u UUIDSlice) Value() (driver.Value, error) {
	if len(u) == 0 {
		return "[]", nil
	}
	return json.Marshal(u)
}

// Scan implements the sql.Scanner interface for UUIDSlice
func (u *UUIDSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal UUIDSlice value: %v", value)
	}
	return json.Unmarshal(bytes, u)
}

// Task represents a task in the system
type Task struct {
	ID             uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Title          string       `json:"title" gorm:"not null"`
	Description    string       `json:"description"`
	Status         TaskStatus   `json:"status" gorm:"not null;default:'Upcoming';index:idx_task_status"`
	Priority       TaskPriority `json:"priority" gorm:"not null;default:'Medium';index:idx_task_priority"`
	CreatedAt      time.Time    `json:"created_at" gorm:"not null;default:current_timestamp"`
	UpdatedAt      time.Time    `json:"updated_at" gorm:"not null;default:current_timestamp"`
	CreatorID      uuid.UUID    `json:"creator_id" gorm:"type:uuid;not null;index:idx_task_creator"`
	AssigneeID     *uuid.UUID   `json:"assignee_id,omitempty" gorm:"type:uuid;index:idx_task_assignee"`
	ReviewerID     *uuid.UUID   `json:"reviewer_id,omitempty" gorm:"type:uuid;"`
	CategoryID     *uuid.UUID   `json:"category_id,omitempty" gorm:"type:uuid"`
	ParentTaskID   *uuid.UUID   `json:"parent_task_id,omitempty" gorm:"type:uuid"`
	ProjectID      uuid.UUID    `json:"project_id" gorm:"type:uuid;not null;index:idx_task_project"`
	OrganizationID uuid.UUID    `json:"organization_id" gorm:"type:uuid;not null;index:idx_task_org"`

	EstimatedHours float64    `json:"estimated_hours,omitempty"`
	ActualHours    float64    `json:"actual_hours,omitempty"`
	StartDate      time.Time  `json:"start_date" gorm:"not null;index:idx_task_dates"`
	Duration       *float64   `json:"duration,omitempty"`
	DueDate        *time.Time `json:"due_date,omitempty" gorm:"index:idx_task_dates"`

	Dependencies    UUIDSlice `json:"dependencies" gorm:"type:jsonb"`
	HealthScore     *float64  `json:"health_score,omitempty"`
	ComplexityScore *float64  `json:"complexity_score,omitempty"`

	// Additional metadata
	AIMetadata      map[string]interface{} `json:"ai_metadata,omitempty" gorm:"type:jsonb"`
	ProgressMetrics map[string]interface{} `json:"progress_metrics,omitempty" gorm:"type:jsonb"`
	Blockers        []string               `json:"blockers,omitempty" gorm:"type:jsonb"`
	RiskFactors     map[string]interface{} `json:"risk_factors,omitempty" gorm:"type:jsonb"`
}

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Title       string    `json:"title" binding:"required" example:"Complete project documentation"`
	Description string    `json:"description" example:"Write comprehensive documentation for the API"`
	Status      string    `json:"status" example:"pending"`
	DueDate     time.Time `json:"due_date,omitempty" example:"2024-12-31T23:59:59Z"`
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Title       string    `json:"title,omitempty" example:"Updated project documentation"`
	Description string    `json:"description,omitempty" example:"Updated documentation details"`
	Status      string    `json:"status,omitempty" example:"in_progress"`
	DueDate     time.Time `json:"due_date,omitempty" example:"2024-12-31T23:59:59Z"`
}

// TaskResponse represents the response for task operations
type TaskResponse struct {
	Task Task `json:"task"`
}

// TaskListResponse represents the response for listing tasks
type TaskListResponse struct {
	Tasks []Task `json:"tasks"`
}

// Common errors
var (
	ErrInvalidStatus  = NewError("invalid task status")
	ErrInvalidCreator = NewError("invalid creator ID")
)

// Error represents a domain error
type Error struct {
	message string
}

// NewError creates a new Error instance
func NewError(message string) *Error {
	return &Error{message: message}
}

// Error returns the error message
func (e *Error) Error() string {
	return e.message
}

func (t TaskStatus) IsValid() bool {
	switch t {
	case TaskStatusUpcoming, TaskStatusInProgress, TaskStatusCompleted,
		TaskStatusCancelled, TaskStatusBlocked, TaskStatusUnderReview,
		TaskStatusDeferred:
		return true
	}
	return false
}

func (t TaskPriority) IsValid() bool {
	switch t {
	case TaskPriorityLow, TaskPriorityMedium, TaskPriorityHigh, TaskPriorityUrgent:
		return true
	}
	return false
}

// TableName specifies the table name for the Task model
func (Task) TableName() string {
	return "tasks"
}

// Validate checks if the task data is valid
func (t *Task) Validate() error {
	if t.Title == "" {
		return ErrInvalidInput
	}
	if !t.Status.IsValid() {
		return ErrInvalidStatus
	}
	if t.CreatorID == uuid.Nil {
		return ErrInvalidCreator
	}
	if t.ProjectID == uuid.Nil {
		return ErrInvalidInput
	}
	if t.OrganizationID == uuid.Nil {
		return ErrInvalidInput
	}
	if !t.Priority.IsValid() {
		return ErrInvalidInput
	}
	return nil
}

// BeforeCreate is called before creating a new task record
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Status == "" {
		t.Status = TaskStatusUpcoming
	}
	if t.Priority == "" {
		t.Priority = TaskPriorityMedium
	}
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return t.Validate()
}

// BeforeUpdate is called before updating a task record
func (t *Task) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return t.Validate()
}

func (t *Task) DashboardUserID() uuid.UUID {
	if t.AssigneeID != nil {
		return *t.AssigneeID
	}
	return t.CreatorID
}
