package user

import (
	"time"

	"github.com/google/uuid"
)

type UserAnalytics struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Action    string    `gorm:"type:varchar(50);not null"`
	Timestamp time.Time `gorm:"not null;default:now()"`
	Metadata  string    `gorm:"type:jsonb"`
}

type SessionAnalytics struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	SessionID  string    `gorm:"type:varchar(64);not null;index"`
	UserID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Action     string    `gorm:"type:varchar(50);not null"`
	DeviceInfo string    `gorm:"type:varchar(255)"`
	IPAddress  string    `gorm:"type:varchar(64)"`
	Timestamp  time.Time `gorm:"not null;default:now()"`
	Metadata   string    `gorm:"type:jsonb"`
}
