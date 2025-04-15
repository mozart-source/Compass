package calendar

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type EventType string

const (
	EventTypeNone     EventType = "None"
	EventTypeTask     EventType = "Task"
	EventTypeMeeting  EventType = "Meeting"
	EventTypeTodo     EventType = "Todo"
	EventTypeHoliday  EventType = "Holiday"
	EventTypeReminder EventType = "Reminder"
)

type RecurrenceType string

const (
	RecurrenceTypeNone     RecurrenceType = "None"
	RecurrenceTypeDaily    RecurrenceType = "Daily"
	RecurrenceTypeWeekly   RecurrenceType = "Weekly"
	RecurrenceTypeBiweekly RecurrenceType = "Biweekly"
	RecurrenceTypeMonthly  RecurrenceType = "Monthly"
	RecurrenceTypeYearly   RecurrenceType = "Yearly"
	RecurrenceTypeCustom   RecurrenceType = "Custom"
)

type OccurrenceStatus string

const (
	OccurrenceStatusUpcoming  OccurrenceStatus = "Upcoming"
	OccurrenceStatusCancelled OccurrenceStatus = "Cancelled"
	OccurrenceStatusCompleted OccurrenceStatus = "Completed"
)

type NotificationMethod string

const (
	NotificationMethodEmail NotificationMethod = "Email"
	NotificationMethodPush  NotificationMethod = "Push"
	NotificationMethodSMS   NotificationMethod = "SMS"
)

type Transparency string

const (
	TransparencyOpaque      Transparency = "opaque"
	TransparencyTransparent Transparency = "transparent"
)

// StringArray represents a PostgreSQL string array type for Swagger documentation
type StringArray []string

// Value implements the driver.Valuer interface
func (a StringArray) Value() (driver.Value, error) {
	return pq.StringArray(a).Value()
}

// Scan implements the sql.Scanner interface
func (a *StringArray) Scan(src interface{}) error {
	return (*pq.StringArray)(a).Scan(src)
}

// Int64Array represents a PostgreSQL integer array type for Swagger documentation
type Int64Array []int64

// Value implements the driver.Valuer interface
func (a Int64Array) Value() (driver.Value, error) {
	return pq.Int64Array(a).Value()
}

// Scan implements the sql.Scanner interface
func (a *Int64Array) Scan(src interface{}) error {
	return (*pq.Int64Array)(a).Scan(src)
}

// EventCollaborator represents a user collaborating on a calendar event (sharing, invitation, permissions)
type EventCollaborator struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	EventID     uuid.UUID  `json:"event_id" gorm:"type:uuid;not null;index:idx_event_collab_event"`
	UserID      uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index:idx_event_collab_user"`
	Role        string     `json:"role" gorm:"type:varchar(50);not null;default:'viewer'"`    // owner, editor, viewer, invitee
	Status      string     `json:"status" gorm:"type:varchar(20);not null;default:'pending'"` // accepted, declined, pending
	InvitedBy   uuid.UUID  `json:"invited_by" gorm:"type:uuid;not null"`
	InvitedAt   time.Time  `json:"invited_at" gorm:"not null;default:current_timestamp"`
	RespondedAt *time.Time `json:"responded_at"`
	CreatedAt   time.Time  `json:"created_at" gorm:"not null;default:current_timestamp"`
	UpdatedAt   time.Time  `json:"updated_at" gorm:"not null;default:current_timestamp"`
}

// CalendarEvent represents a calendar event or series
type CalendarEvent struct {
	ID           uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID       uuid.UUID    `json:"user_id" gorm:"type:uuid;not null;index:idx_calendar_event_user"`
	Title        string       `json:"title" gorm:"type:varchar(255);not null;index:idx_calendar_event_title"`
	Description  string       `json:"description" gorm:"type:text"`
	EventType    EventType    `json:"event_type" gorm:"type:varchar(50);not null;default:'None'"`
	StartTime    time.Time    `json:"start_time" gorm:"not null;index:idx_calendar_event_start"`
	EndTime      time.Time    `json:"end_time" gorm:"not null;index:idx_calendar_event_end"`
	IsAllDay     bool         `json:"is_all_day" gorm:"not null;default:false"`
	Location     string       `json:"location,omitempty" gorm:"type:varchar(255)"`
	Color        string       `json:"color,omitempty" gorm:"type:varchar(7)"`
	Transparency Transparency `json:"transparency" gorm:"type:varchar(20);not null;default:'opaque'"`
	CreatedAt    time.Time    `json:"created_at" gorm:"not null;default:current_timestamp"`
	UpdatedAt    time.Time    `json:"updated_at" gorm:"not null;default:current_timestamp"`

	// Relationships (for preload fun)
	RecurrenceRules []RecurrenceRule     `json:"recurrence_rules,omitempty" gorm:"foreignKey:EventID"`
	Occurrences     []OccurrenceResponse `json:"occurrences,omitempty" gorm:"-"` // Use - to exclude from DB operations
	Exceptions      []EventException     `json:"exceptions,omitempty" gorm:"foreignKey:EventID"`
	Reminders       []EventReminder      `json:"reminders,omitempty" gorm:"foreignKey:EventID"`
	Collaborators   []EventCollaborator  `json:"collaborators,omitempty" gorm:"foreignKey:EventID"`
}

// RecurrenceRule represents the recurrence pattern for a calendar event
type RecurrenceRule struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	EventID    uuid.UUID      `json:"event_id" gorm:"type:uuid;not null;index:idx_recurrence_event"`
	Freq       RecurrenceType `json:"freq" gorm:"type:varchar(50);not null;default:'None'"`
	Interval   int            `json:"interval" gorm:"not null;default:1"`
	ByDay      StringArray    `json:"by_day,omitempty" gorm:"type:varchar[]"`
	ByMonth    Int64Array     `json:"by_month,omitempty" gorm:"type:integer[]"`
	ByMonthDay Int64Array     `json:"by_month_day,omitempty" gorm:"type:integer[]"`
	Count      *int           `json:"count,omitempty"`
	Until      *time.Time     `json:"until,omitempty"`
	CreatedAt  time.Time      `json:"created_at" gorm:"not null;default:current_timestamp"`
	UpdatedAt  time.Time      `json:"updated_at" gorm:"not null;default:current_timestamp"`
}

// EventOccurrence represents a single instance of a recurring event
type EventOccurrence struct {
	ID             uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	EventID        uuid.UUID        `json:"event_id" gorm:"type:uuid;not null;index:idx_occurrence_event"`
	OccurrenceTime time.Time        `json:"occurrence_time" gorm:"not null;index:idx_occurrence_time"`
	Status         OccurrenceStatus `json:"status" gorm:"type:varchar(50);not null;default:'Upcoming'"`
	CreatedAt      time.Time        `json:"created_at" gorm:"not null;default:current_timestamp"`
	UpdatedAt      time.Time        `json:"updated_at" gorm:"not null;default:current_timestamp"`
}

// OccurrenceResponse represents an occurrence with its overridden values
type OccurrenceResponse struct {
	EventOccurrence
	Title        *string       `json:"title,omitempty"`
	Description  *string       `json:"description,omitempty"`
	Location     *string       `json:"location,omitempty"`
	Color        *string       `json:"color,omitempty"`
	Transparency *Transparency `json:"transparency,omitempty"`
	EndTime      *time.Time    `json:"end_time,omitempty"`
}

// EventException represents modifications to a specific occurrence
type EventException struct {
	ID                   uuid.UUID     `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	EventID              uuid.UUID     `json:"event_id" gorm:"type:uuid;not null;index:idx_exception_event"`
	OriginalTime         time.Time     `json:"original_time" gorm:"not null;index:idx_exception_time"`
	OccurrenceID         uuid.UUID     `json:"occurrence_id,omitempty" gorm:"type:uuid;index:idx_exception_occurrence"`
	IsDeleted            bool          `json:"is_deleted" gorm:"not null;default:false"`
	OverrideStartTime    *time.Time    `json:"override_start_time,omitempty"`
	OverrideEndTime      *time.Time    `json:"override_end_time,omitempty"`
	OverrideTitle        *string       `json:"override_title,omitempty" gorm:"type:varchar(255)"`
	OverrideDescription  *string       `json:"override_description,omitempty" gorm:"type:text"`
	OverrideLocation     *string       `json:"override_location,omitempty" gorm:"type:varchar(255)"`
	OverrideColor        *string       `json:"override_color,omitempty" gorm:"type:varchar(7)"`
	OverrideTransparency *Transparency `json:"override_transparency,omitempty" gorm:"type:varchar(20)"`
	CreatedAt            time.Time     `json:"created_at" gorm:"not null;default:current_timestamp"`
	UpdatedAt            time.Time     `json:"updated_at" gorm:"not null;default:current_timestamp"`
}

// EventReminder represents a reminder for an event
type EventReminder struct {
	ID            uuid.UUID          `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	EventID       uuid.UUID          `json:"event_id" gorm:"type:uuid;not null;index:idx_reminder_event"`
	MinutesBefore int                `json:"minutes_before" gorm:"not null"`
	Method        NotificationMethod `json:"method" gorm:"type:varchar(50);not null"`
	CreatedAt     time.Time          `json:"created_at" gorm:"not null;default:current_timestamp"`
	UpdatedAt     time.Time          `json:"updated_at" gorm:"not null;default:current_timestamp"`
}

// TableName specifies the table names for each model
func (CalendarEvent) TableName() string     { return "calendar_events" }
func (RecurrenceRule) TableName() string    { return "recurrence_rules" }
func (EventOccurrence) TableName() string   { return "event_occurrences" }
func (EventException) TableName() string    { return "event_exceptions" }
func (EventReminder) TableName() string     { return "event_reminders" }
func (EventCollaborator) TableName() string { return "event_collaborators" }

// BeforeCreate hooks for UUID generation
func (e *CalendarEvent) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (r *RecurrenceRule) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

func (o *EventOccurrence) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return nil
}

func (e *EventException) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (r *EventReminder) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// BeforeCreate hook for EventCollaborator
func (c *EventCollaborator) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	if c.InvitedAt.IsZero() {
		c.InvitedAt = time.Now()
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return nil
}

// Request/Response DTOs
type CreateCalendarEventRequest struct {
	Title        string       `json:"title" binding:"required"`
	Description  string       `json:"description"`
	EventType    EventType    `json:"event_type" binding:"required"`
	StartTime    time.Time    `json:"start_time" binding:"required"`
	EndTime      time.Time    `json:"end_time" binding:"required"`
	IsAllDay     bool         `json:"is_all_day"`
	Location     string       `json:"location"`
	Color        string       `json:"color"`
	Transparency Transparency `json:"transparency"`

	// Optional recurrence
	RecurrenceRule *CreateRecurrenceRuleRequest `json:"recurrence_rule,omitempty"`
	// Optional reminders
	Reminders []CreateEventReminderRequest `json:"reminders,omitempty"`
}

type CreateRecurrenceRuleRequest struct {
	Freq       RecurrenceType `json:"freq" binding:"required"`
	Interval   int            `json:"interval" binding:"required,min=1"`
	ByDay      []string       `json:"by_day,omitempty"`
	ByMonth    []int          `json:"by_month,omitempty"`
	ByMonthDay []int          `json:"by_month_day,omitempty"`
	Count      *int           `json:"count,omitempty"`
	Until      *time.Time     `json:"until,omitempty"`
}

type CreateEventReminderRequest struct {
	MinutesBefore int                `json:"minutes_before" binding:"required,min=0"`
	Method        NotificationMethod `json:"method" binding:"required"`
}

type UpdateCalendarEventRequest struct {
	Title                *string       `json:"title,omitempty"`
	Description          *string       `json:"description,omitempty"`
	EventType            *EventType    `json:"event_type,omitempty"`
	StartTime            *time.Time    `json:"start_time,omitempty"`
	EndTime              *time.Time    `json:"end_time,omitempty"`
	IsAllDay             *bool         `json:"is_all_day,omitempty"`
	Location             *string       `json:"location,omitempty"`
	Color                *string       `json:"color,omitempty"`
	Transparency         *Transparency `json:"transparency,omitempty"`
	PreserveDateSequence *bool         `json:"preserve_date_sequence,omitempty"`
}

type CalendarEventResponse struct {
	Event       CalendarEvent        `json:"event"`
	Occurrences []OccurrenceResponse `json:"occurrences,omitempty"`
}

type CalendarEventListResponse struct {
	Events []CalendarEvent `json:"events"`
	Total  int64           `json:"total"`
}

// Common errors
var (
	ErrInvalidEventType    = NewError("invalid event type")
	ErrInvalidTimeRange    = NewError("end time must be after start time")
	ErrInvalidRecurrence   = NewError("invalid recurrence configuration")
	ErrInvalidReminderTime = NewError("invalid reminder time")
	ErrInvalidTransparency = NewError("invalid transparency value")
)

// Error type
type Error struct {
	message string
}

func NewError(message string) *Error {
	return &Error{message: message}
}

func (e *Error) Error() string {
	return e.message
}

// Validation methods
func (e *CalendarEvent) Validate() error {
	if e.Title == "" {
		return NewError("title is required")
	}
	if e.StartTime.After(e.EndTime) {
		return ErrInvalidTimeRange
	}
	if !isValidEventType(e.EventType) {
		return ErrInvalidEventType
	}
	if !isValidTransparency(e.Transparency) {
		return ErrInvalidTransparency
	}
	return nil
}

func (r *RecurrenceRule) Validate() error {
	if !isValidRecurrenceType(r.Freq) {
		return ErrInvalidRecurrence
	}
	if r.Interval < 1 {
		return NewError("interval must be at least 1")
	}
	if r.Count != nil && *r.Count < 1 {
		return NewError("count must be at least 1")
	}
	if r.Until != nil && r.Until.Before(time.Now()) {
		return NewError("until date must be in the future")
	}
	return nil
}

func (r *EventReminder) Validate() error {
	if r.MinutesBefore < 0 {
		return ErrInvalidReminderTime
	}
	if !isValidNotificationMethod(r.Method) {
		return NewError("invalid notification method")
	}
	return nil
}

// Helper functions for validation
func isValidEventType(t EventType) bool {
	switch t {
	case EventTypeNone, EventTypeTask, EventTypeMeeting, EventTypeTodo,
		EventTypeHoliday, EventTypeReminder:
		return true
	}
	return false
}

func isValidRecurrenceType(t RecurrenceType) bool {
	switch t {
	case RecurrenceTypeNone, RecurrenceTypeDaily, RecurrenceTypeWeekly,
		RecurrenceTypeBiweekly, RecurrenceTypeMonthly, RecurrenceTypeYearly,
		RecurrenceTypeCustom:
		return true
	}
	return false
}

func isValidNotificationMethod(m NotificationMethod) bool {
	switch m {
	case NotificationMethodEmail, NotificationMethodPush, NotificationMethodSMS:
		return true
	}
	return false
}

func isValidTransparency(t Transparency) bool {
	switch t {
	case TransparencyOpaque, TransparencyTransparent:
		return true
	}
	return false
}
