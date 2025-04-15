package habits

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Habit struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null"`
	Title             string     `gorm:"size:255;not null"`
	Description       string     `gorm:"type:text"`
	StartDay          time.Time  `gorm:"not null;default:current_timestamp"`
	EndDay            *time.Time `gorm:"default:null"`
	CurrentStreak     int        `gorm:"default:0;not null"`
	StreakStartDate   *time.Time `gorm:"default:null"`
	LongestStreak     int        `gorm:"default:0;not null"`
	IsCompleted       bool       `gorm:"default:false;not null"`
	LastCompletedDate *time.Time `gorm:"default:null"`
	CreatedAt         time.Time  `gorm:"not null;default:current_timestamp"`
	UpdatedAt         time.Time  `gorm:"not null;default:current_timestamp;autoUpdateTime"`
	StreakQuality     float64    `gorm:"default:0;not null"` // Stored in DB for faster retrieval
}

// StreakHistory represents a historical record of a habit streak
type StreakHistory struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	HabitID       uuid.UUID `gorm:"type:uuid;not null"`
	StartDate     time.Time `gorm:"not null"`
	EndDate       time.Time `gorm:"not null"`
	StreakLength  int       `gorm:"not null"`
	CompletedDays int       `gorm:"not null"`
	CreatedAt     time.Time `gorm:"not null;default:current_timestamp"`
}

// CreateHabitInput represents the input for creating a new habit
type CreateHabitInput struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	StartDay    time.Time  `json:"start_day"`
	EndDay      *time.Time `json:"end_day"`
	UserID      uuid.UUID  `json:"user_id"`
}

// UpdateHabitInput represents the input for updating a habit
type UpdateHabitInput struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	StartDay    *time.Time `json:"start_day,omitempty"`
	EndDay      *time.Time `json:"end_day,omitempty"`
}

// HabitResponse represents the response body for a habit
type HabitResponse struct {
	Habit Habit `json:"habit"`
}

// HabitListResponse represents the response body for a list of habits
type HabitListResponse struct {
	Habits []Habit `json:"habits"`
}

// TableName specifies the table name for the Habit model
func (Habit) TableName() string {
	return "habits"
}

// HabitCompletionLog represents a record of each habit completion
type HabitCompletionLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	HabitID   uuid.UUID `gorm:"type:uuid;not null;index:idx_habit_completion,priority:1"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_habit_completion,priority:2;index:idx_user_date,priority:1"`
	Date      time.Time `gorm:"not null;index:idx_habit_completion,priority:3;index:idx_user_date,priority:2"`
	CreatedAt time.Time `gorm:"not null;default:current_timestamp"`
}

// TableName specifies the table name for the HabitCompletionLog model
func (HabitCompletionLog) TableName() string {
	return "habit_completion_logs"
}

// BeforeCreate is called before creating a new habit record
func (h *Habit) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	h.CreatedAt = time.Now()
	h.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate is called before updating a habit record
func (h *Habit) BeforeUpdate(tx *gorm.DB) error {
	h.UpdatedAt = time.Now()
	return nil
}
