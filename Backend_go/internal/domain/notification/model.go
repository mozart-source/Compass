package notification

import (
	"time"

	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// Type represents the type of notification
type Type string

const (
	// Notification types
	General      = "general"
	UserMention  = "user_mention"
	NewComment   = "new_comment"
	NewLike      = "new_like"
	TaskAssigned = "task_assigned"
	Reminder     = "reminder"
	System       = "system"

	// Habit notification types
	HabitCompleted = "habit_completed"
	HabitStreak    = "habit_streak"
	HabitBroken    = "habit_broken"
	HabitReminder  = "habit_reminder"
	HabitMilestone = "habit_milestone"

	// Event collaboration notification types
	EventInvite            = "event_invite"
	EventInviteAccepted    = "event_invite_accepted"
	EventInviteDeclined    = "event_invite_declined"
	EventRemovedFromCollab = "event_removed_from_collab"

	// Workflow notification types
	WorkflowActionRequired = "workflow_action_required"
	WorkflowApproved       = "workflow_approved"
	WorkflowRejected       = "workflow_rejected"
	WorkflowCompleted      = "workflow_completed"
	WorkflowFailed         = "workflow_failed"
)

// Status represents the status of a notification
type Status string

const (
	// Unread status for new notifications
	Unread Status = "UNREAD"
	// Read status for viewed notifications
	Read Status = "READ"
	// Archived status for archived notifications
	Archived Status = "ARCHIVED"
)

// StringMap is a type for storing string-to-string mappings in JSONB fields
type StringMap map[string]string

func (m *StringMap) Scan(value interface{}) error {
	if value == nil {
		*m = make(map[string]string)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSONB value: %v", value)
	}

	// Create a new map for unmarshaling
	result := make(map[string]string)
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}

	// Assign the result to the receiver
	*m = result
	return nil
}

func (m StringMap) Value() (driver.Value, error) {
	if m == nil || len(m) == 0 {
		return []byte("{}"), nil
	}
	return json.Marshal(m)
}

// Notification represents a notification entity
type Notification struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	Type        Type       `json:"type" gorm:"not null"`
	Title       string     `json:"title" gorm:"not null"`
	Content     string     `json:"content" gorm:"not null"`
	Status      Status     `json:"status" gorm:"not null;default:'UNREAD'"`
	Data        StringMap  `json:"data" gorm:"type:jsonb"`
	Reference   string     `json:"reference" gorm:"index"`
	ReferenceID uuid.UUID  `json:"reference_id" gorm:"type:uuid;index"`
	CreatedAt   time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"not null"`
	ReadAt      *time.Time `json:"read_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// BeforeCreate hook to set default values
func (n *Notification) BeforeCreate() error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	if n.UpdatedAt.IsZero() {
		n.UpdatedAt = time.Now()
	}
	if n.Status == "" {
		n.Status = Unread
	}
	return nil
}
