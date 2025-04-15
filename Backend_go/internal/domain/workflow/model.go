package workflow

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type WorkflowStatus string

const (
	WorkflowStatusPending     WorkflowStatus = "pending"
	WorkflowStatusActive      WorkflowStatus = "active"
	WorkflowStatusPaused      WorkflowStatus = "paused"
	WorkflowStatusCompleted   WorkflowStatus = "completed"
	WorkflowStatusFailed      WorkflowStatus = "failed"
	WorkflowStatusCancelled   WorkflowStatus = "cancelled"
	WorkflowStatusArchived    WorkflowStatus = "archived"
	WorkflowStatusUnderReview WorkflowStatus = "under_review"
	WorkflowStatusOptimizing  WorkflowStatus = "optimizing"
)

type WorkflowType string

const (
	WorkflowTypeSequential  WorkflowType = "sequential"
	WorkflowTypeParallel    WorkflowType = "parallel"
	WorkflowTypeConditional WorkflowType = "conditional"
	WorkflowTypeAIDriven    WorkflowType = "ai_driven"
	WorkflowTypeHybrid      WorkflowType = "hybrid"
)

// Workflow represents a workflow in the system
type Workflow struct {
	ID             uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name           string         `json:"name" gorm:"type:varchar(255);not null"`
	Description    string         `json:"description" gorm:"type:text"`
	WorkflowType   WorkflowType   `json:"workflow_type" gorm:"type:varchar(50);not null;default:'sequential'"`
	CreatedBy      uuid.UUID      `json:"created_by" gorm:"type:uuid;not null"`
	OrganizationID uuid.UUID      `json:"organization_id" gorm:"type:uuid;not null"`
	Status         WorkflowStatus `json:"status" gorm:"type:varchar(50);not null;default:'pending'"`

	// Configuration & Metadata
	Config           datatypes.JSON `json:"config" gorm:"type:jsonb;default:'{}'"`
	WorkflowMetadata datatypes.JSON `json:"workflow_metadata" gorm:"type:jsonb;default:'{}'"`
	Version          string         `json:"version" gorm:"type:varchar(50)"`
	Tags             pq.StringArray `json:"tags" gorm:"type:text[]"`

	// AI Integration
	AIEnabled             bool           `json:"ai_enabled" gorm:"default:false"`
	AIConfidenceThreshold float64        `json:"ai_confidence_threshold" gorm:"default:0.8"`
	AIOverrideRules       datatypes.JSON `json:"ai_override_rules" gorm:"type:jsonb;default:'{}'"`
	AILearningData        datatypes.JSON `json:"ai_learning_data" gorm:"type:jsonb;default:'{}'"`

	// Performance Metrics
	AverageCompletionTime float64        `json:"average_completion_time" gorm:"default:0.0"`
	SuccessRate           float64        `json:"success_rate" gorm:"default:0.0"`
	OptimizationScore     float64        `json:"optimization_score" gorm:"default:0.0"`
	BottleneckAnalysis    datatypes.JSON `json:"bottleneck_analysis" gorm:"type:jsonb;default:'{}'"`

	// Time Management
	EstimatedDuration   *int           `json:"estimated_duration"`
	ActualDuration      *int           `json:"actual_duration"`
	ScheduleConstraints datatypes.JSON `json:"schedule_constraints" gorm:"type:jsonb;default:'{}'"`
	Deadline            *time.Time     `json:"deadline"`

	// Error Handling
	ErrorHandlingConfig datatypes.JSON `json:"error_handling_config" gorm:"type:jsonb;default:'{}'"`
	RetryPolicy         datatypes.JSON `json:"retry_policy" gorm:"type:jsonb;default:'{}'"`
	FallbackSteps       datatypes.JSON `json:"fallback_steps" gorm:"type:jsonb;default:'{}'"`

	// Audit & Compliance
	ComplianceRules datatypes.JSON `json:"compliance_rules" gorm:"type:jsonb;default:'{}'"`
	AuditTrail      datatypes.JSON `json:"audit_trail" gorm:"type:jsonb;default:'{}'"`
	AccessControl   datatypes.JSON `json:"access_control" gorm:"type:jsonb;default:'{}'"`

	// Timestamps
	CreatedAt        time.Time  `json:"created_at" gorm:"not null;default:current_timestamp"`
	UpdatedAt        time.Time  `json:"updated_at" gorm:"not null;default:current_timestamp"`
	LastExecutedAt   *time.Time `json:"last_executed_at"`
	NextScheduledRun *time.Time `json:"next_scheduled_run"`
}

// CreateWorkflowRequest represents the request body for creating a workflow
type CreateWorkflowRequest struct {
	Name              string         `json:"name" binding:"required" example:"New Project Workflow"`
	Description       string         `json:"description" example:"Workflow for managing new project creation"`
	WorkflowType      WorkflowType   `json:"workflow_type" binding:"required" example:"sequential"`
	OrganizationID    uuid.UUID      `json:"organization_id" binding:"required"`
	Config            datatypes.JSON `json:"config,omitempty"`
	AIEnabled         bool           `json:"ai_enabled,omitempty"`
	EstimatedDuration *int           `json:"estimated_duration,omitempty"`
	Deadline          *time.Time     `json:"deadline,omitempty"`
	Tags              pq.StringArray `json:"tags,omitempty"`
}

// UpdateWorkflowRequest represents the request body for updating a workflow
type UpdateWorkflowRequest struct {
	Name              *string         `json:"name,omitempty"`
	Description       *string         `json:"description,omitempty"`
	Status            *WorkflowStatus `json:"status,omitempty"`
	Config            datatypes.JSON  `json:"config,omitempty"`
	AIEnabled         *bool           `json:"ai_enabled,omitempty"`
	EstimatedDuration *int            `json:"estimated_duration,omitempty"`
	Deadline          *time.Time      `json:"deadline,omitempty"`
	Tags              pq.StringArray  `json:"tags,omitempty"`
}

// WorkflowResponse represents the response for workflow operations
type WorkflowResponse struct {
	Workflow *Workflow `json:"workflow"`
}

// WorkflowListResponse represents the response for listing workflows
type WorkflowListResponse struct {
	Workflows []Workflow `json:"workflows"`
	Total     int64      `json:"total"`
}

// TableName specifies the table name for the Workflow model
func (Workflow) TableName() string {
	return "workflows"
}

// BeforeCreate is called before creating a new workflow record
func (w *Workflow) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	if w.Status == "" {
		w.Status = WorkflowStatusPending
	}
	if w.WorkflowType == "" {
		w.WorkflowType = WorkflowTypeSequential
	}
	w.CreatedAt = time.Now()
	w.UpdatedAt = time.Now()
	return nil
}

// BeforeUpdate is called before updating a workflow record
func (w *Workflow) BeforeUpdate(tx *gorm.DB) error {
	w.UpdatedAt = time.Now()
	return nil
}

// WorkflowFilter represents the filter options for querying workflows
type WorkflowFilter struct {
	OrganizationID *uuid.UUID
	CreatedBy      *uuid.UUID
	Status         *WorkflowStatus
	WorkflowType   *WorkflowType
	StartDate      *time.Time
	EndDate        *time.Time
	Tags           []string
	Page           int
	PageSize       int
}

// IsValid checks if the workflow type is valid
func (wt WorkflowType) IsValid() bool {
	switch wt {
	case WorkflowTypeSequential, WorkflowTypeParallel, WorkflowTypeConditional,
		WorkflowTypeAIDriven, WorkflowTypeHybrid:
		return true
	}
	return false
}

// IsValid checks if the workflow status is valid
func (ws WorkflowStatus) IsValid() bool {
	switch ws {
	case WorkflowStatusPending, WorkflowStatusActive, WorkflowStatusPaused,
		WorkflowStatusCompleted, WorkflowStatusFailed, WorkflowStatusCancelled,
		WorkflowStatusArchived, WorkflowStatusUnderReview, WorkflowStatusOptimizing:
		return true
	}
	return false
}
