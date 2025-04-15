package dto

import (
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/calendar"
	"github.com/google/uuid"
)

// Calendar Event DTOs
type CreateCalendarEventRequest struct {
	Title        string                `json:"title" binding:"required"`
	Description  string                `json:"description"`
	EventType    calendar.EventType    `json:"event_type" binding:"required"`
	StartTime    time.Time             `json:"start_time" binding:"required"`
	EndTime      time.Time             `json:"end_time" binding:"required"`
	IsAllDay     bool                  `json:"is_all_day"`
	Location     string                `json:"location"`
	Color        string                `json:"color" binding:"omitempty,len=7"` // e.g. "#ff8000"
	Transparency calendar.Transparency `json:"transparency"`

	// Optional fields for recurring events
	RecurrenceRule *CreateRecurrenceRuleRequest `json:"recurrence_rule,omitempty"`
	Reminders      []CreateEventReminderRequest `json:"reminders,omitempty"`
}

type CreateRecurrenceRuleRequest struct {
	Freq       calendar.RecurrenceType `json:"freq" binding:"required"`
	Interval   int                     `json:"interval" binding:"required,min=1"`
	ByDay      []string                `json:"by_day,omitempty" binding:"omitempty,dive,oneof=MO TU WE TH FR SA SU"`
	ByMonth    []int                   `json:"by_month,omitempty" binding:"omitempty,dive,min=1,max=12"`
	ByMonthDay []int                   `json:"by_month_day,omitempty" binding:"omitempty,dive,min=1,max=31"`
	Count      *int                    `json:"count,omitempty" binding:"omitempty,min=1"`
	Until      *time.Time              `json:"until,omitempty"`
}

type CreateEventReminderRequest struct {
	MinutesBefore int                         `json:"minutes_before" binding:"required,min=0"`
	Method        calendar.NotificationMethod `json:"method" binding:"required"`
}

type UpdateCalendarEventRequest struct {
	Title        *string                `json:"title,omitempty"`
	Description  *string                `json:"description,omitempty"`
	EventType    *calendar.EventType    `json:"event_type,omitempty"`
	StartTime    *time.Time             `json:"start_time,omitempty"`
	EndTime      *time.Time             `json:"end_time,omitempty"`
	IsAllDay     *bool                  `json:"is_all_day,omitempty"`
	Location     *string                `json:"location,omitempty"`
	Color        *string                `json:"color,omitempty" binding:"omitempty,len=7"`
	Transparency *calendar.Transparency `json:"transparency,omitempty"`
}

type UpdateOccurrenceRequest struct {
	EventID      uuid.UUID                  `json:"event_id" binding:"required"`
	OriginalTime time.Time                  `json:"original_time" binding:"required"`
	Updates      UpdateCalendarEventRequest `json:"updates" binding:"required"`
}

// Response DTOs
type CalendarEventResponse struct {
	Event       calendar.CalendarEvent     `json:"event"`
	Occurrences []calendar.EventOccurrence `json:"occurrences,omitempty"`
	Exceptions  []calendar.EventException  `json:"exceptions,omitempty"`
}

type CalendarEventListResponse struct {
	Events []calendar.CalendarEvent `json:"events"`
	Total  int64                    `json:"total"`
	Page   int                      `json:"page"`
	Size   int                      `json:"size"`
}

type OccurrenceListResponse struct {
	Occurrences []calendar.EventOccurrence `json:"occurrences"`
	Total       int64                      `json:"total"`
}

// Query parameters
type ListEventsParams struct {
	StartTime time.Time           `form:"start_time" binding:"required"`
	EndTime   time.Time           `form:"end_time" binding:"required"`
	EventType *calendar.EventType `form:"event_type"`
	Page      int                 `form:"page,default=1" binding:"min=1"`
	PageSize  int                 `form:"page_size,default=10" binding:"min=1,max=100"`
	Search    string              `form:"search"`
}

// Collaboration DTOs

type InviteCollaboratorRequest struct {
	EventID uuid.UUID `json:"event_id" binding:"required"`
	UserID  uuid.UUID `json:"user_id" binding:"required"`
	Role    string    `json:"role" binding:"required"`
}

type RespondToInviteRequest struct {
	EventID uuid.UUID `json:"event_id" binding:"required"`
	Accept  bool      `json:"accept" binding:"required"`
}

type CollaboratorResponse struct {
	ID          uuid.UUID  `json:"id"`
	EventID     uuid.UUID  `json:"event_id"`
	UserID      uuid.UUID  `json:"user_id"`
	Role        string     `json:"role"`
	Status      string     `json:"status"`
	InvitedBy   uuid.UUID  `json:"invited_by"`
	InvitedAt   time.Time  `json:"invited_at"`
	RespondedAt *time.Time `json:"responded_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ListCollaboratorsResponse struct {
	Collaborators []CollaboratorResponse `json:"collaborators"`
}

type RemoveCollaboratorRequest struct {
	EventID uuid.UUID `json:"event_id" binding:"required"`
	UserID  uuid.UUID `json:"user_id" binding:"required"`
}
