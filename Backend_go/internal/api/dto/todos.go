package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateTodoRequest struct {
	Title                 string                 `json:"title" binding:"required"`
	Description           string                 `json:"description"`
	Status                string                 `json:"status" binding:"required"`
	Priority              string                 `json:"priority" binding:"required"`
	DueDate               *time.Time             `json:"due_date"`
	ReminderTime          *time.Time             `json:"reminder_time"`
	IsRecurring           bool                   `json:"is_recurring"`
	RecurrencePattern     map[string]interface{} `json:"recurrence_pattern"`
	Tags                  map[string]interface{} `json:"tags"`
	Checklist             map[string]interface{} `json:"checklist"`
	LinkedTaskID          *uuid.UUID             `json:"linked_task_id"`
	LinkedCalendarEventID *uuid.UUID             `json:"linked_calendar_event_id"`
	UserID                uuid.UUID              `json:"user_id"`
	ListID                *uuid.UUID             `json:"list_id,omitempty"`
	IsCompleted           bool                   `json:"is_completed"`
	CompletedAt           *time.Time             `json:"completed_at"`
}

type UpdateTodoRequest struct {
	Title                 *string                 `json:"title,omitempty"`
	Description           *string                 `json:"description,omitempty"`
	Status                *string                 `json:"status,omitempty"`
	Priority              *string                 `json:"priority,omitempty"`
	DueDate               *time.Time              `json:"due_date,omitempty"`
	ReminderTime          *time.Time              `json:"reminder_time,omitempty"`
	IsRecurring           *bool                   `json:"is_recurring,omitempty"`
	RecurrencePattern     *map[string]interface{} `json:"recurrence_pattern,omitempty"`
	Tags                  *map[string]interface{} `json:"tags,omitempty"`
	Checklist             *map[string]interface{} `json:"checklist,omitempty"`
	LinkedTaskID          *uuid.UUID              `json:"linked_task_id,omitempty"`
	LinkedCalendarEventID *uuid.UUID              `json:"linked_calendar_event_id,omitempty"`
	IsCompleted           *bool                   `json:"is_completed,omitempty"`
	CompletedAt           *time.Time              `json:"completed_at,omitempty"`
}

type TodoResponse struct {
	ID                    uuid.UUID              `json:"id"`
	Title                 string                 `json:"title"`
	Description           string                 `json:"description"`
	Status                string                 `json:"status"`
	Priority              string                 `json:"priority"`
	DueDate               *time.Time             `json:"due_date"`
	ReminderTime          *time.Time             `json:"reminder_time"`
	IsRecurring           bool                   `json:"is_recurring"`
	RecurrencePattern     map[string]interface{} `json:"recurrence_pattern"`
	Tags                  map[string]interface{} `json:"tags"`
	Checklist             map[string]interface{} `json:"checklist"`
	LinkedTaskID          *uuid.UUID             `json:"linked_task_id"`
	LinkedCalendarEventID *uuid.UUID             `json:"linked_calendar_event_id"`
	IsCompleted           bool                   `json:"is_completed"`
	CompletedAt           *time.Time             `json:"completed_at"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
	UserID                uuid.UUID              `json:"user_id"`
	ListID                uuid.UUID              `json:"list_id"`
}

type TodoListResponse struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	IsDefault   bool            `json:"is_default"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	UserID      uuid.UUID       `json:"user_id"`
	Todos       []*TodoResponse `json:"todos"`
	TotalCount  int64           `json:"total_count"`
	Page        int             `json:"page"`
	PageSize    int             `json:"page_size"`
}

type TodoListsResponse struct {
	Lists []TodoListResponse `json:"lists"`
}

type TodoFilterRequest struct {
	Status                string    `form:"status" example:"In Progress"`
	Priority              string    `form:"priority" example:"High"`
	IsCompleted           bool      `form:"is_completed" example:"true"`
	DueDate               time.Time `form:"due_date" example:"2024-03-15T09:00:00Z"`
	ReminderTime          time.Time `form:"reminder_time" example:"2024-03-15T09:00:00Z"`
	IsRecurring           bool      `form:"is_recurring" example:"true"`
	Tags                  []string  `form:"tags" example:"tag1,tag2"`
	Checklist             []string  `form:"checklist" example:"item1,item2"`
	LinkedTaskID          uuid.UUID `form:"linked_task_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	LinkedCalendarEventID uuid.UUID `form:"linked_calendar_event_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Page                  int       `form:"page" example:"1"`
	PageSize              int       `form:"page_size" example:"20"`
}

type UserTodosResponse struct {
	Todos      []*TodoResponse `json:"todos"`
	TotalCount int64           `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
}

type UpdateTodoStatusRequest struct {
	Status string `json:"status" binding:"required" example:"In Progress"`
}

type UpdateTodoPriorityRequest struct {
	Priority string `json:"priority" binding:"required" example:"High"`
}
