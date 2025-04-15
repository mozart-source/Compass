package task

import (
	"time"

	"github.com/google/uuid"
)

type TaskAnalytics struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	TaskID    uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Action    string    `gorm:"type:varchar(50);not null"`
	Timestamp time.Time `gorm:"not null;default:now()"`
	Metadata  string    `gorm:"type:jsonb"`
}
