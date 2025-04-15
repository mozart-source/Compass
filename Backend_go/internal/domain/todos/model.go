package todos

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TodoPriority represents the priority level of a todo
type TodoPriority string

const (
	PriorityHigh   TodoPriority = "high"
	PriorityMedium TodoPriority = "medium"
	PriorityLow    TodoPriority = "low"
)

// TodoStatus represents the status of a todo
type TodoStatus string

const (
	StatusPending    TodoStatus = "pending"
	StatusInProgress TodoStatus = "in_progress"
	StatusArchived   TodoStatus = "archived"
)

// TodoList represents a collection of todos
type TodoList struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Name        string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text"`
	IsDefault   bool      `gorm:"default:false;not null"`
	CreatedAt   time.Time `gorm:"not null;default:current_timestamp"`
	UpdatedAt   time.Time `gorm:"not null;default:current_timestamp"`
	Todos       []Todo    `gorm:"foreignKey:ListID"`
}

// Todo represents a todo item in the system
type Todo struct {
	ID                    uuid.UUID    `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID                uuid.UUID    `gorm:"type:uuid;not null;index"`
	ListID                uuid.UUID    `gorm:"type:uuid;not null;index"` // Reference to TodoList
	Title                 string       `gorm:"size:255;not null"`
	Description           string       `gorm:"type:text"`
	Status                TodoStatus   `gorm:"type:varchar(20);not null;default:'pending';index"`
	Priority              TodoPriority `gorm:"type:varchar(20);not null;default:'medium';index"`
	IsCompleted           bool         `gorm:"not null;default:false;index"`
	CompletionDate        *time.Time
	DueDate               *time.Time `gorm:"index"`
	ReminderTime          *time.Time
	IsRecurring           bool                   `gorm:"default:false;not null"`
	RecurrencePattern     map[string]interface{} `gorm:"type:jsonb;default:'{}';serializer:json"`
	Tags                  map[string]interface{} `gorm:"type:jsonb;default:'{}';serializer:json"`
	Checklist             map[string]interface{} `gorm:"type:jsonb;default:'{}';serializer:json"`
	LinkedTaskID          *uuid.UUID             `gorm:"type:uuid"`
	LinkedCalendarEventID *uuid.UUID             `gorm:"type:uuid"`
	AIGenerated           bool                   `gorm:"default:false;not null"`
	AISuggestions         map[string]interface{} `gorm:"type:jsonb;default:'{}';serializer:json"`
	CreatedAt             time.Time              `gorm:"not null;default:current_timestamp;index"`
	UpdatedAt             time.Time              `gorm:"not null;default:current_timestamp;autoUpdateTime"`
	List                  TodoList               `gorm:"foreignKey:ListID"` // Relationship to TodoList
}

// CreateTodoRequest represents the request body for creating a todo
type CreateTodoRequest struct {
	Title                 string                 `json:"title"`
	Description           string                 `json:"description"`
	Status                TodoStatus             `json:"status"`
	Priority              TodoPriority           `json:"priority"`
	DueDate               *time.Time             `json:"due_date"`
	ReminderTime          *time.Time             `json:"reminder_time"`
	IsRecurring           bool                   `json:"is_recurring"`
	RecurrencePattern     map[string]interface{} `json:"recurrence_pattern"`
	Tags                  map[string]interface{} `json:"tags"`
	Checklist             map[string]interface{} `json:"checklist"`
	LinkedTaskID          *uuid.UUID             `json:"linked_task_id"`
	LinkedCalendarEventID *uuid.UUID             `json:"linked_calendar_event_id"`
	UserID                uuid.UUID              `json:"user_id"`
	ListID                uuid.UUID              `json:"list_id"`
}

// UpdateTodoRequest represents the request body for updating a todo
type UpdateTodoRequest struct {
	Title                 *string                `json:"title,omitempty"`
	Description           *string                `json:"description,omitempty"`
	Status                *TodoStatus            `json:"status,omitempty"`
	Priority              *TodoPriority          `json:"priority,omitempty"`
	DueDate               *time.Time             `json:"due_date,omitempty"`
	ReminderTime          *time.Time             `json:"reminder_time,omitempty"`
	IsRecurring           *bool                  `json:"is_recurring,omitempty"`
	RecurrencePattern     map[string]interface{} `json:"recurrence_pattern,omitempty"`
	Tags                  map[string]interface{} `json:"tags,omitempty"`
	Checklist             map[string]interface{} `json:"checklist,omitempty"`
	LinkedTaskID          *uuid.UUID             `json:"linked_task_id,omitempty"`
	LinkedCalendarEventID *uuid.UUID             `json:"linked_calendar_event_id,omitempty"`
	UserID                uuid.UUID              `json:"user_id"`
	ListID                uuid.UUID              `json:"list_id"`
}

// TodoResponse represents the response body for a todo
type TodoResponse struct {
	Todo Todo `json:"todo"`
}

// TodoListResponse represents a single todo list in responses
type TodoListResponse struct {
	List TodoList `json:"list"`
}

// TodoListsResponse represents multiple todo lists in responses
type TodoListsResponse struct {
	Lists []TodoList `json:"lists"`
}

// Common errors
var (
	ErrInvalidStatus   = NewError("invalid todo status")
	ErrInvalidPriority = NewError("invalid todo priority")
	ErrInvalidInput    = NewError("invalid input")
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

func (s TodoStatus) IsValid() bool {
	switch s {
	case StatusPending, StatusInProgress, StatusArchived:
		return true
	}
	return false
}

func (p TodoPriority) IsValid() bool {
	switch p {
	case PriorityHigh, PriorityMedium, PriorityLow:
		return true
	}
	return false
}

func (Todo) TableName() string {
	return "todos"
}

func (TodoList) TableName() string {
	return "todo_lists"
}

// Validate checks if the todo data is valid
func (t *Todo) Validate() error {
	if t.Title == "" {
		return ErrInvalidInput
	}
	if !t.Status.IsValid() {
		return ErrInvalidStatus
	}
	if !t.Priority.IsValid() {
		return ErrInvalidPriority
	}
	if t.UserID == uuid.Nil {
		return ErrInvalidInput
	}
	if t.ListID == uuid.Nil {
		return ErrInvalidInput
	}
	return nil
}

// BeforeCreate is called before creating a new todo list
func (l *TodoList) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}

	l.CreatedAt = time.Now()
	l.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate is called before updating a todo list
func (l *TodoList) BeforeUpdate(tx *gorm.DB) error {
	l.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate is called before creating a new todo record
func (t *Todo) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}

	// Set default values if not provided
	if t.Status == "" {
		t.Status = StatusPending
	}
	if t.Priority == "" {
		t.Priority = PriorityMedium
	}

	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return t.Validate()
}

// BeforeUpdate is called before updating a todo record
func (t *Todo) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()

	// If marked as completed and completion date is not set
	if t.IsCompleted && t.CompletionDate == nil {
		now := time.Now()
		t.CompletionDate = &now
	}

	return t.Validate()
}

type TodoFilter struct {
	UserID                *uuid.UUID
	Status                *TodoStatus
	Priority              *TodoPriority
	IsCompleted           *bool
	DueDateStart          *time.Time
	DueDateEnd            *time.Time
	DueDate               *time.Time
	ReminderTime          *time.Time
	IsRecurring           *bool
	Tags                  *[]string
	Checklist             *[]string
	LinkedTaskID          *uuid.UUID
	LinkedCalendarEventID *uuid.UUID
	Page                  int
	PageSize              int
}
