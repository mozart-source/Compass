package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateTaskRequest represents the request body for creating a task
// @Description Request body for creating a new task in the system
type CreateTaskRequest struct {
	Title          string      `json:"title" binding:"required"`
	Description    string      `json:"description"`
	Status         string      `json:"status" binding:"required"`
	Priority       string      `json:"priority" binding:"required"`
	ProjectID      uuid.UUID   `json:"project_id" binding:"required"`
	OrganizationID uuid.UUID   `json:"organization_id" binding:"required"`
	AssigneeID     *uuid.UUID  `json:"assignee_id,omitempty"`
	ReviewerID     *uuid.UUID  `json:"reviewer_id,omitempty"`
	CategoryID     *uuid.UUID  `json:"category_id,omitempty"`
	ParentTaskID   *uuid.UUID  `json:"parent_task_id,omitempty"`
	EstimatedHours float64     `json:"estimated_hours,omitempty"`
	StartDate      time.Time   `json:"start_date" binding:"required"`
	Duration       *float64    `json:"duration,omitempty"`
	DueDate        *time.Time  `json:"due_date,omitempty"`
	Dependencies   []uuid.UUID `json:"dependencies,omitempty"`
}

// UpdateTaskRequest represents the request body for updating a task
// @Description Request body for updating task information
type UpdateTaskRequest struct {
	Title          *string     `json:"title,omitempty"`
	Description    *string     `json:"description,omitempty"`
	Status         *string     `json:"status,omitempty"`
	Priority       *string     `json:"priority,omitempty"`
	AssigneeID     *uuid.UUID  `json:"assignee_id,omitempty"`
	ReviewerID     *uuid.UUID  `json:"reviewer_id,omitempty"`
	CategoryID     *uuid.UUID  `json:"category_id,omitempty"`
	EstimatedHours *float64    `json:"estimated_hours,omitempty"`
	StartDate      *time.Time  `json:"start_date,omitempty"`
	Duration       *float64    `json:"duration,omitempty"`
	DueDate        *time.Time  `json:"due_date,omitempty"`
	Dependencies   []uuid.UUID `json:"dependencies,omitempty"`
}

// TaskResponse represents a task in API responses
// @Description Detailed task information returned in API responses
type TaskResponse struct {
	ID             uuid.UUID  `json:"id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CreatorID      uuid.UUID  `json:"creator_id"`
	AssigneeID     *uuid.UUID `json:"assignee_id,omitempty"`
	ReviewerID     *uuid.UUID `json:"reviewer_id,omitempty"`
	CategoryID     *uuid.UUID `json:"category_id,omitempty"`
	ParentTaskID   *uuid.UUID `json:"parent_task_id,omitempty"`
	ProjectID      uuid.UUID  `json:"project_id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	EstimatedHours float64    `json:"estimated_hours,omitempty"`
	StartDate      time.Time  `json:"start_date"`
	Duration       *float64   `json:"duration,omitempty"`
	DueDate        *time.Time `json:"due_date,omitempty"`
}

// TaskListResponse represents a paginated list of tasks with metadata
type TaskListResponse struct {
	Tasks      []TaskResponse `json:"tasks"`
	TotalCount int64          `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

// TaskFilterRequest represents the query parameters for filtering tasks
type TaskFilterRequest struct {
	OrganizationID string    `form:"organization_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProjectID      string    `form:"project_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Status         string    `form:"status" example:"In Progress"`
	Priority       string    `form:"priority" example:"High"`
	AssigneeID     string    `form:"assignee_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	CreatorID      string    `form:"creator_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	ReviewerID     string    `form:"reviewer_id" example:"550e8400-e29b-41d4-a716-446655440004"`
	StartDate      time.Time `form:"start_date" example:"2024-03-15T09:00:00Z"`
	EndDate        time.Time `form:"end_date" example:"2024-03-20T17:00:00Z"`
	Page           int       `form:"page" example:"1"`
	PageSize       int       `form:"page_size" example:"20"`
}

// UpdateTaskStatusRequest represents the request body for updating a task's status
type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required" example:"In Progress"`
}

// AssignTaskRequest represents the request body for assigning a task to a user
type AssignTaskRequest struct {
	AssigneeID string `json:"assignee_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000"`
}
