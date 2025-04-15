package workflow

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WorkflowStepExecution represents the execution of a workflow step
type WorkflowStepExecution struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	ExecutionID       uuid.UUID      `json:"execution_id" gorm:"type:uuid;not null"`
	StepID            uuid.UUID      `json:"step_id" gorm:"type:uuid;not null"`
	Status            StepStatus     `json:"status" gorm:"type:varchar(50);not null;default:'pending'"`
	ExecutionPriority int            `json:"execution_priority" gorm:"default:0"`
	ExecutionMetadata datatypes.JSON `json:"execution_metadata" gorm:"type:jsonb"`
	StartedAt         time.Time      `json:"started_at" gorm:"not null;default:current_timestamp"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"not null;default:current_timestamp"`
	CompletedAt       *time.Time     `json:"completed_at"`
	Result            datatypes.JSON `json:"result" gorm:"type:jsonb"`
	Error             *string        `json:"error"`
}

// WorkflowExecution represents the execution of a workflow
type WorkflowExecution struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	WorkflowID        uuid.UUID      `json:"workflow_id" gorm:"type:uuid;not null"`
	Status            WorkflowStatus `json:"status" gorm:"type:varchar(50);not null;default:'pending'"`
	ExecutionPriority int            `json:"execution_priority" gorm:"default:0"`
	ExecutionMetadata datatypes.JSON `json:"execution_metadata" gorm:"type:jsonb"`
	StartedAt         time.Time      `json:"started_at" gorm:"not null;default:current_timestamp"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"not null;default:current_timestamp"`
	CompletedAt       *time.Time     `json:"completed_at"`
	Result            datatypes.JSON `json:"result" gorm:"type:jsonb"`
	Error             *string        `json:"error"`
}

// WorkflowAgentLink represents a link between a workflow and an agent
type WorkflowAgentLink struct {
	ID                  uuid.UUID      `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	WorkflowID          uuid.UUID      `json:"workflow_id" gorm:"type:uuid;not null"`
	AgentID             uuid.UUID      `json:"agent_id" gorm:"type:uuid;not null"`
	InteractionType     string         `json:"interaction_type" gorm:"type:varchar(100);not null"`
	ConfidenceScore     *float64       `json:"confidence_score"`
	InteractionMetadata datatypes.JSON `json:"interaction_metadata" gorm:"type:jsonb"`
	CreatedAt           time.Time      `json:"created_at" gorm:"not null;default:current_timestamp"`
}

// CreateWorkflowExecutionRequest represents the request body for creating a workflow execution
type CreateWorkflowExecutionRequest struct {
	WorkflowID        uuid.UUID      `json:"workflow_id" binding:"required"`
	ExecutionPriority *int           `json:"execution_priority,omitempty"`
	ExecutionMetadata datatypes.JSON `json:"execution_metadata,omitempty"`
}

// UpdateWorkflowExecutionRequest represents the request body for updating a workflow execution
type UpdateWorkflowExecutionRequest struct {
	Status            *WorkflowStatus `json:"status,omitempty"`
	ExecutionPriority *int            `json:"execution_priority,omitempty"`
	ExecutionMetadata datatypes.JSON  `json:"execution_metadata,omitempty"`
	Result            datatypes.JSON  `json:"result,omitempty"`
	Error             *string         `json:"error,omitempty"`
}

// WorkflowExecutionResponse represents the response for execution operations
type WorkflowExecutionResponse struct {
	Execution      *WorkflowExecution      `json:"execution"`
	StepExecutions []WorkflowStepExecution `json:"step_executions,omitempty"`
}

// WorkflowExecutionListResponse represents the response for listing executions
type WorkflowExecutionListResponse struct {
	Executions []WorkflowExecution `json:"executions"`
	Total      int64               `json:"total"`
}

// TableName specifies the table names for each model
func (WorkflowStepExecution) TableName() string {
	return "workflow_step_executions"
}

func (WorkflowExecution) TableName() string {
	return "workflow_executions"
}

func (WorkflowAgentLink) TableName() string {
	return "workflow_agent_links"
}

// BeforeCreate hooks for UUID generation and timestamps
func (e *WorkflowStepExecution) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	e.StartedAt = time.Now()
	e.UpdatedAt = time.Now()
	return nil
}

func (e *WorkflowExecution) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	e.StartedAt = time.Now()
	e.UpdatedAt = time.Now()
	return nil
}

func (l *WorkflowAgentLink) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	l.CreatedAt = time.Now()
	return nil
}

// BeforeUpdate hooks for timestamps
func (e *WorkflowStepExecution) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return nil
}

func (e *WorkflowExecution) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return nil
}

// WorkflowExecutionFilter represents the filter options for querying workflow executions
type WorkflowExecutionFilter struct {
	WorkflowID *uuid.UUID
	Status     *WorkflowStatus
	StartDate  *time.Time
	EndDate    *time.Time
	Page       int
	PageSize   int
}
