package workflow

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusActive    StepStatus = "active"
	StepStatusCompleted StepStatus = "completed"
	StepStatusSkipped   StepStatus = "skipped"
	StepStatusFailed    StepStatus = "failed"
)

type StepType string

const (
	StepTypeManual       StepType = "manual"
	StepTypeAutomated    StepType = "automated"
	StepTypeApproval     StepType = "approval"
	StepTypeNotification StepType = "notification"
	StepTypeIntegration  StepType = "integration"
	StepTypeDecision     StepType = "decision"
	StepTypeAITask       StepType = "ai_task"
)

// WorkflowStep represents a step in a workflow
type WorkflowStep struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	WorkflowID  uuid.UUID  `json:"workflow_id" gorm:"type:uuid;not null"`
	Name        string     `json:"name" gorm:"type:varchar(255);not null"`
	Description string     `json:"description" gorm:"type:text"`
	StepType    StepType   `json:"step_type" gorm:"type:varchar(50);not null"`
	StepOrder   int        `json:"step_order" gorm:"not null"`
	Status      StepStatus `json:"status" gorm:"type:varchar(50);not null;default:'pending'"`

	// Configuration
	Config     datatypes.JSON `json:"config" gorm:"type:jsonb"`
	Conditions datatypes.JSON `json:"conditions" gorm:"type:jsonb"`
	Timeout    *int           `json:"timeout"`

	// Execution Control
	RetryConfig  datatypes.JSON `json:"retry_config" gorm:"type:jsonb"`
	IsRequired   bool           `json:"is_required" gorm:"default:true"`
	AutoAdvance  bool           `json:"auto_advance" gorm:"default:false"`
	CanRevert    bool           `json:"can_revert" gorm:"default:false"`
	Dependencies []string       `json:"dependencies" gorm:"type:text[]"`

	// Version Control
	Version           string     `json:"version" gorm:"type:varchar(50);default:'1.0.0'"`
	PreviousVersionID *uuid.UUID `json:"previous_version_id" gorm:"type:uuid"`

	// Performance Metrics
	AverageExecutionTime float64        `json:"average_execution_time" gorm:"default:0.0"`
	SuccessRate          float64        `json:"success_rate" gorm:"default:0.0"`
	LastExecutionResult  datatypes.JSON `json:"last_execution_result" gorm:"type:jsonb"`

	// Assignment & Notifications
	AssignedTo         *uuid.UUID     `json:"assigned_to" gorm:"type:uuid"`
	AssignedToRoleID   *uuid.UUID     `json:"assigned_to_role_id" gorm:"type:uuid"`
	NotificationConfig datatypes.JSON `json:"notification_config" gorm:"type:jsonb"`

	CreatedAt time.Time `json:"created_at" gorm:"not null;default:current_timestamp"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null;default:current_timestamp"`
}

// CreateWorkflowStepRequest represents the request body for creating a workflow step
type CreateWorkflowStepRequest struct {
	Name             string         `json:"name" binding:"required"`
	Description      string         `json:"description"`
	StepType         StepType       `json:"step_type" binding:"required"`
	StepOrder        int            `json:"step_order" binding:"required"`
	Config           datatypes.JSON `json:"config,omitempty"`
	Conditions       datatypes.JSON `json:"conditions,omitempty"`
	Timeout          *int           `json:"timeout,omitempty"`
	IsRequired       *bool          `json:"is_required,omitempty"`
	AssignedTo       *uuid.UUID     `json:"assigned_to,omitempty"`
	AssignedToRoleID *uuid.UUID     `json:"assigned_to_role_id,omitempty"`
}

// UpdateWorkflowStepRequest represents the request body for updating a workflow step
type UpdateWorkflowStepRequest struct {
	Name             *string        `json:"name,omitempty"`
	Description      *string        `json:"description,omitempty"`
	StepType         *StepType      `json:"step_type,omitempty"`
	StepOrder        *int           `json:"step_order,omitempty"`
	Status           *StepStatus    `json:"status,omitempty"`
	Config           datatypes.JSON `json:"config,omitempty"`
	Conditions       datatypes.JSON `json:"conditions,omitempty"`
	Timeout          *int           `json:"timeout,omitempty"`
	IsRequired       *bool          `json:"is_required,omitempty"`
	AssignedTo       *uuid.UUID     `json:"assigned_to,omitempty"`
	AssignedToRoleID *uuid.UUID     `json:"assigned_to_role_id,omitempty"`
}

// WorkflowStepResponse represents the response for step operations
type WorkflowStepResponse struct {
	Step *WorkflowStep `json:"step"`
}

// WorkflowStepListResponse represents the response for listing steps
type WorkflowStepListResponse struct {
	Steps []WorkflowStep `json:"steps"`
	Total int64          `json:"total"`
}

// TableName specifies the table name for the WorkflowStep model
func (WorkflowStep) TableName() string {
	return "workflow_steps"
}

// BeforeCreate is called before creating a new workflow step record
func (s *WorkflowStep) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	if s.Status == "" {
		s.Status = StepStatusPending
	}
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate is called before updating a workflow step record
func (s *WorkflowStep) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}

// WorkflowStepFilter represents the filter options for querying workflow steps
type WorkflowStepFilter struct {
	WorkflowID *uuid.UUID
	StepType   *StepType
	Status     *StepStatus
	AssignedTo *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
}
