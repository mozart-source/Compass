package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateUserRequest represents the request body for user registration
// @Description Request body for creating a new user
type CreateUserRequest struct {
	Email          string     `json:"email" binding:"required,email" example:"user@example.com"`
	Username       string     `json:"username" binding:"required" example:"johndoe"`
	Password       string     `json:"password" binding:"required,min=8" example:"securePass123"`
	FirstName      string     `json:"first_name" binding:"required" example:"John"`
	LastName       string     `json:"last_name" binding:"required" example:"Doe"`
	PhoneNumber    string     `json:"phone_number" example:"+1234567890"`
	Timezone       string     `json:"timezone" example:"UTC"`
	Locale         string     `json:"locale" example:"en-US"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
}

// UpdateUserRequest represents the request body for updating user profile
// @Description Request body for updating user profile information
type UpdateUserRequest struct {
	Email                   *string                `json:"email,omitempty" binding:"omitempty,email" example:"newemail@example.com"`
	FirstName               *string                `json:"first_name,omitempty" example:"John"`
	LastName                *string                `json:"last_name,omitempty" example:"Doe"`
	PhoneNumber             *string                `json:"phone_number,omitempty" example:"+1234567890"`
	AvatarURL               *string                `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	Bio                     *string                `json:"bio,omitempty" example:"Software Engineer"`
	Timezone                *string                `json:"timezone,omitempty" example:"UTC"`
	Locale                  *string                `json:"locale,omitempty" example:"en-US"`
	NotificationPreferences map[string]interface{} `json:"notification_preferences,omitempty"`
	WorkspaceSettings       map[string]interface{} `json:"workspace_settings,omitempty"`
	Username                *string                `json:"username,omitempty" example:"johndoe"`
}

// UserResponse represents the user data returned in API responses
// @Description User information returned in API responses
type UserResponse struct {
	ID          uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email       string     `json:"email" example:"user@example.com"`
	Username    string     `json:"username" example:"johndoe"`
	IsActive    bool       `json:"is_active" example:"true"`
	IsSuperuser bool       `json:"is_superuser" example:"false"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	FirstName   string     `json:"first_name,omitempty" example:"John"`
	LastName    string     `json:"last_name,omitempty" example:"Doe"`
	PhoneNumber string     `json:"phone_number,omitempty" example:"+1234567890"`
	AvatarURL   string     `json:"avatar_url,omitempty" example:"https://example.com/avatar.jpg"`
	Bio         string     `json:"bio,omitempty" example:"Software Engineer"`
	Timezone    string     `json:"timezone,omitempty" example:"UTC"`
	Locale      string     `json:"locale,omitempty" example:"en-US"`

	// Security info
	MFAEnabled          bool       `json:"mfa_enabled" example:"false"`
	LastLogin           *time.Time `json:"last_login,omitempty"`
	FailedLoginAttempts int        `json:"failed_login_attempts" example:"0"`
	AccountLockedUntil  *time.Time `json:"account_locked_until,omitempty"`
	ForcePasswordChange bool       `json:"force_password_change" example:"false"`

	// Organization
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`

	// Preferences
	NotificationPreferences map[string]interface{} `json:"notification_preferences,omitempty"`
	AllowedIPRanges         []string               `json:"allowed_ip_ranges,omitempty"`
	MaxSessions             int                    `json:"max_sessions" example:"5"`
	WorkspaceSettings       map[string]interface{} `json:"workspace_settings,omitempty"`
}

// UserListResponse represents a paginated list of users
// @Description Paginated list of users with metadata
type UserListResponse struct {
	Users      []UserResponse `json:"users"`
	TotalCount int64          `json:"total_count" example:"100"`
	Page       int            `json:"page" example:"1"`
	PageSize   int            `json:"page_size" example:"20"`
}

// LoginRequest represents the request body for user login
// @Description Request body for user authentication
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"securePass123"`
}

// LoginResponse represents the response after successful login
// @Description Response containing authentication token and user information
type LoginResponse struct {
	Token     string          `json:"token"`
	User      UserResponse    `json:"user"`
	Session   SessionResponse `json:"session"`
	ExpiresAt time.Time       `json:"expires_at"`
}

// TokenResponse represents a JWT token response
// @Description JWT token information
type TokenResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string `json:"token_type" example:"Bearer"`
	ExpiresIn   int    `json:"expires_in" example:"3600"`
}
