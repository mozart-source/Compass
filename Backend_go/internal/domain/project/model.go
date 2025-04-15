package project

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Common errors
var (
	ErrProjectNotFound   = errors.New("project not found")
	ErrInvalidInput      = errors.New("invalid input")
	ErrProjectNameExists = errors.New("project name already exists in organization")
)

type ProjectStatus string

const (
	ProjectStatusActive    ProjectStatus = "Active"
	ProjectStatusCompleted ProjectStatus = "Completed"
	ProjectStatusArchived  ProjectStatus = "Archived"
	ProjectStatusOnHold    ProjectStatus = "On Hold"
)

// IsValid validates the project status
func (s ProjectStatus) IsValid() bool {
	switch s {
	case ProjectStatusActive, ProjectStatusCompleted, ProjectStatusArchived, ProjectStatusOnHold:
		return true
	}
	return false
}

type Project struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name           string         `json:"name" gorm:"type:varchar(255);not null;uniqueIndex:idx_project_name,where:deleted_at is null"`
	Description    string         `json:"description" gorm:"type:text"`
	Status         ProjectStatus  `json:"status" gorm:"not null;default:'active';index:idx_project_status"`
	OrganizationID uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null;uniqueIndex:idx_project_name_org,priority:1"`
	CreatorID      uuid.UUID      `json:"creator_id" gorm:"type:uuid;not null;index:idx_project_creator"`
	OwnerID        uuid.UUID      `json:"owner_id" gorm:"type:uuid;not null;index:idx_project_owner"`
	StartDate      time.Time      `json:"start_date" gorm:"not null;index:idx_project_dates"`
	EndDate        *time.Time     `json:"end_date,omitempty" gorm:"index:idx_project_dates"`
	CreatedAt      time.Time      `json:"created_at" gorm:"index:idx_project_created"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// BeforeCreate is called before inserting a new project
func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	if p.Status == "" {
		p.Status = ProjectStatusActive
	}
	if !p.Status.IsValid() {
		return errors.New("invalid project status")
	}
	return nil
}

// BeforeUpdate is called before updating a project
func (p *Project) BeforeUpdate(tx *gorm.DB) error {
	if !p.Status.IsValid() {
		return errors.New("invalid project status")
	}
	return nil
}

type CreateProjectInput struct {
	Name           string        `validate:"required,min=3,max=100"`
	Description    string        `validate:"max=500"`
	Status         ProjectStatus `validate:"required,oneof=active inactive archived"`
	OrganizationID uuid.UUID     `validate:"required"`
	CreatorID      uuid.UUID     `validate:"required"`
	OwnerID        uuid.UUID     `validate:"required"`
	StartDate      time.Time     `validate:"required"`
	EndDate        *time.Time    `validate:"omitempty"`
}

type UpdateProjectInput struct {
	Name        *string        `validate:"omitempty,min=3,max=100"`
	Description *string        `validate:"omitempty,max=500"`
	Status      *ProjectStatus `validate:"omitempty,oneof=active inactive archived"`
	OwnerID     *uuid.UUID     `validate:"omitempty"`
	StartDate   *time.Time     `validate:"omitempty"`
	EndDate     *time.Time     `validate:"omitempty"`
}

type ProjectFilter struct {
	Page           int            `validate:"min=0"`
	PageSize       int            `validate:"min=1,max=100"`
	Name           *string        `validate:"omitempty,max=100"`
	Status         *ProjectStatus `validate:"omitempty,oneof=active inactive archived"`
	OrganizationID *uuid.UUID     `validate:"required"`
}

type ProjectMember struct {
	UserID   uuid.UUID `json:"user_id"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type ProjectDetails struct {
	Project      *Project        `json:"project"`
	MembersCount int64           `json:"members_count"`
	TasksCount   int64           `json:"tasks_count"`
	Members      []ProjectMember `json:"members"`
}

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name           string        `json:"name" binding:"required" example:"New Project"`
	Description    string        `json:"description" example:"A detailed project description"`
	Status         ProjectStatus `json:"status" example:"Active"`
	OrganizationID uuid.UUID     `json:"organization_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}
