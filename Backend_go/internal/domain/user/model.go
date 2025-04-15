package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserStatus represents the status of a user
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusBlocked  UserStatus = "blocked"
)

// User represents a user in the system
type User struct {
	ID                  uuid.UUID              `json:"id" gorm:"type:uuid;primary_key"`
	Email               string                 `json:"email" gorm:"uniqueIndex:idx_user_email,where:deleted_at is null;not null"`
	Username            string                 `json:"username" gorm:"uniqueIndex:idx_user_username,where:deleted_at is null;not null"`
	FirstName           string                 `json:"first_name" gorm:"not null"`
	LastName            string                 `json:"last_name" gorm:"not null"`
	PhoneNumber         string                 `json:"phone_number"`
	AvatarURL           string                 `json:"avatar_url"`
	Bio                 string                 `json:"bio"`
	Timezone            string                 `json:"timezone" gorm:"not null;default:'GMT+2'"`
	Locale              string                 `json:"locale" gorm:"not null;default:'en-US'"`
	PasswordHash        string                 `json:"-" gorm:"not null"`
	Status              UserStatus             `json:"status" gorm:"not null;default:'active'"`
	IsActive            bool                   `json:"is_active" gorm:"default:true;index:idx_user_active"`
	IsSuperuser         bool                   `json:"is_superuser" gorm:"default:false;index:idx_user_superuser"`
	CreatedAt           time.Time              `json:"created_at" gorm:"index:idx_user_created"`
	UpdatedAt           time.Time              `json:"updated_at"`
	DeletedAt           *time.Time             `json:"deleted_at,omitempty" gorm:"index"`
	MFAEnabled          bool                   `json:"mfa_enabled" gorm:"default:false"`
	MFASecret           string                 `json:"-"`
	MFABackupCodes      []string               `json:"-" gorm:"-"`                       // Not stored directly in DB
	MFABackupCodesHash  string                 `json:"-" gorm:"column:mfa_backup_codes"` // Stored as JSON string
	FailedLoginAttempts int                    `json:"-" gorm:"default:0"`
	AccountLockedUntil  *time.Time             `json:"-" gorm:"index:idx_user_locked"`
	Preferences         map[string]interface{} `json:"preferences,omitempty" gorm:"type:jsonb"`
	Provider            string                 `json:"provider,omitempty" gorm:"index:idx_user_provider"`
	ProviderID          string                 `json:"provider_id,omitempty" gorm:"index:idx_user_provider_id"`
	ProviderData        map[string]interface{} `json:"provider_data,omitempty" gorm:"type:jsonb"`
}

// CreateUserRequest represents the request body for user registration
type CreateUserRequest struct {
	Email       string                 `json:"email" binding:"required,email" example:"user@example.com"`
	Username    string                 `json:"username" binding:"required" example:"johndoe"`
	Password    string                 `json:"password" binding:"required" example:"securepassword123"`
	FirstName   string                 `json:"first_name" binding:"required" example:"John"`
	LastName    string                 `json:"last_name" binding:"required" example:"Doe"`
	PhoneNumber string                 `json:"phone_number,omitempty" example:"+1234567890"`
	AvatarURL   string                 `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	Bio         string                 `json:"bio,omitempty" example:"Software developer"`
	Timezone    string                 `json:"timezone,omitempty" example:"GMT+2"`
	Locale      string                 `json:"locale,omitempty" example:"en-US"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"securepassword123"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Username    string                 `json:"username,omitempty" example:"johndoe_updated"`
	Email       string                 `json:"email,omitempty" example:"updated@example.com"`
	Password    string                 `json:"password,omitempty" example:"newpassword123"`
	FirstName   string                 `json:"first_name,omitempty" example:"John"`
	LastName    string                 `json:"last_name,omitempty" example:"Doe"`
	PhoneNumber string                 `json:"phone_number,omitempty" example:"+1234567890"`
	AvatarURL   string                 `json:"avatar_url,omitempty" example:"https://example.com/avatar_updated.jpg"`
	Bio         string                 `json:"bio,omitempty" example:"Updated bio"`
	Timezone    string                 `json:"timezone,omitempty" example:"GMT+3"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

// UserResponse represents the response for user operations
type UserResponse struct {
	User User `json:"user"`
}

// LoginResponse represents the response for successful login
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  User   `json:"user"`
}

// UserListResponse represents the response for listing users
type UserListResponse struct {
	Users []User `json:"users"`
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// BeforeCreate is called before creating a new user record
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate is called before updating a user record
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}
