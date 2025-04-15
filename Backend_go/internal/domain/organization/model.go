package organization

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// OrganizationStatus represents the status of an organization
type OrganizationStatus string

const (
	// OrganizationStatusActive represents an active organization
	OrganizationStatusActive OrganizationStatus = "Active"
	// OrganizationStatusInactive represents an inactive organization
	OrganizationStatusInactive OrganizationStatus = "Inactive"
	// OrganizationStatusArchived represents an archived organization
	OrganizationStatusArchived OrganizationStatus = "Archived"
	// OrganizationStatusSuspended represents a suspended organization
	OrganizationStatusSuspended OrganizationStatus = "Suspended"
)

// IsValid checks if the organization status is valid
func (s OrganizationStatus) IsValid() bool {
	switch s {
	case OrganizationStatusActive, OrganizationStatusInactive,
		OrganizationStatusArchived, OrganizationStatusSuspended:
		return true
	default:
		return false
	}
}

// Organization represents an organization entity in the system
type Organization struct {
	ID          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string                 `json:"name" gorm:"type:varchar(255);not null;uniqueIndex:idx_org_name,where:deleted_at is null"`
	Description string                 `json:"description" gorm:"type:text"`
	Status      OrganizationStatus     `json:"status" gorm:"type:varchar(20);not null;default:'Active';index:idx_org_status"`
	CreatedAt   time.Time              `json:"created_at" gorm:"not null;default:current_timestamp;index:idx_org_created"`
	UpdatedAt   time.Time              `json:"updated_at" gorm:"not null;default:current_timestamp"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty" gorm:"index"`
	CreatorID   uuid.UUID              `json:"creator_id" gorm:"type:uuid;not null;index:idx_org_creator"`
	OwnerID     uuid.UUID              `json:"owner_id" gorm:"type:uuid;not null"`
	Settings    map[string]interface{} `json:"settings,omitempty" gorm:"type:jsonb"`
	Preferences map[string]interface{} `json:"preferences,omitempty" gorm:"type:jsonb"`
}

// TableName specifies the table name for the Organization model
func (Organization) TableName() string {
	return "organizations"
}

// Validate checks if the organization data is valid
func (o *Organization) Validate() error {
	if o.Name == "" {
		return ErrInvalidInput
	}
	if !o.Status.IsValid() {
		return ErrInvalidStatus
	}
	if o.CreatorID == uuid.Nil {
		return ErrInvalidCreator
	}
	if o.OwnerID == uuid.Nil {
		return ErrInvalidOwner
	}
	return nil
}

// BeforeCreate is called before creating a new organization record
func (o *Organization) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	if o.Status == "" {
		o.Status = OrganizationStatusActive
	}
	o.CreatedAt = time.Now()
	o.UpdatedAt = time.Now()
	return o.Validate()
}

// BeforeUpdate is called before updating an organization record
func (o *Organization) BeforeUpdate(tx *gorm.DB) error {
	o.UpdatedAt = time.Now()
	return o.Validate()
}

// Common errors
var (
	ErrOrganizationNotFound = NewError("organization not found")
	ErrInvalidInput         = NewError("invalid input")
	ErrInvalidStatus        = NewError("invalid organization status")
	ErrDuplicateName        = NewError("organization name already exists")
	ErrInvalidCreator       = NewError("invalid creator ID")
	ErrInvalidOwner         = NewError("invalid owner ID")
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
