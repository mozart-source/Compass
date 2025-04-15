package workflow

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Repository defines the interface for workflow data operations
type Repository interface {
	// Workflow operations
	Create(ctx context.Context, workflow *Workflow) error
	Update(ctx context.Context, workflow *Workflow) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Workflow, error)
	List(ctx context.Context, filter *WorkflowFilter) ([]Workflow, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status WorkflowStatus) error

	// Step operations
	CreateStep(ctx context.Context, step *WorkflowStep) error
	UpdateStep(ctx context.Context, step *WorkflowStep) error
	DeleteStep(ctx context.Context, id uuid.UUID) error
	GetStepByID(ctx context.Context, id uuid.UUID) (*WorkflowStep, error)
	ListSteps(ctx context.Context, filter *WorkflowStepFilter) ([]WorkflowStep, int64, error)
	UpdateStepStatus(ctx context.Context, id uuid.UUID, status StepStatus) error

	// Transition operations
	CreateTransition(ctx context.Context, transition *WorkflowTransition) error
	UpdateTransition(ctx context.Context, transition *WorkflowTransition) error
	DeleteTransition(ctx context.Context, id uuid.UUID) error
	GetTransitionByID(ctx context.Context, id uuid.UUID) (*WorkflowTransition, error)
	ListTransitions(ctx context.Context, filter *WorkflowTransitionFilter) ([]WorkflowTransition, int64, error)

	// Execution operations
	CreateExecution(ctx context.Context, execution *WorkflowExecution) error
	UpdateExecution(ctx context.Context, execution *WorkflowExecution) error
	GetExecutionByID(ctx context.Context, id uuid.UUID) (*WorkflowExecution, error)
	ListExecutions(ctx context.Context, filter *WorkflowExecutionFilter) ([]WorkflowExecution, int64, error)
	CancelActiveExecutions(ctx context.Context, workflowID uuid.UUID) error

	// Step execution operations
	CreateStepExecution(ctx context.Context, execution *WorkflowStepExecution) error
	UpdateStepExecution(ctx context.Context, execution *WorkflowStepExecution) error
	GetStepExecutionByID(ctx context.Context, id uuid.UUID) (*WorkflowStepExecution, error)
	ListStepExecutions(ctx context.Context, executionID uuid.UUID) ([]WorkflowStepExecution, error)

	// Agent link operations
	CreateAgentLink(ctx context.Context, link *WorkflowAgentLink) error
	GetAgentLinksByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]WorkflowAgentLink, error)

	// CreateWorkflow creates a new workflow
	CreateWorkflow(ctx context.Context, workflow *Workflow) error
}

// repository implements the Repository interface
type repository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewRepository creates a new workflow repository
func NewRepository(db *gorm.DB, logger *logrus.Logger) Repository {
	return &repository{db: db, logger: logger}
}

// Workflow operations
func (r *repository) Create(ctx context.Context, workflow *Workflow) error {
	return r.db.WithContext(ctx).Create(workflow).Error
}

func (r *repository) Update(ctx context.Context, workflow *Workflow) error {
	return r.db.WithContext(ctx).Save(workflow).Error
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&Workflow{}, id).Error
}

func (r *repository) GetByID(ctx context.Context, id uuid.UUID) (*Workflow, error) {
	var workflow Workflow
	err := r.db.WithContext(ctx).First(&workflow, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

func (r *repository) List(ctx context.Context, filter *WorkflowFilter) ([]Workflow, int64, error) {
	var workflows []Workflow
	var total int64

	query := r.db.WithContext(ctx).Model(&Workflow{})

	if filter != nil {
		if filter.OrganizationID != nil {
			query = query.Where("organization_id = ?", filter.OrganizationID)
		}
		if filter.CreatedBy != nil {
			query = query.Where("created_by = ?", filter.CreatedBy)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.WorkflowType != nil {
			query = query.Where("workflow_type = ?", filter.WorkflowType)
		}
		if filter.StartDate != nil {
			query = query.Where("created_at >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("created_at <= ?", filter.EndDate)
		}
		if len(filter.Tags) > 0 {
			query = query.Where("tags && ?", filter.Tags)
		}
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err = query.Find(&workflows).Error
	if err != nil {
		return nil, 0, err
	}

	return workflows, total, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id uuid.UUID, status WorkflowStatus) error {
	return r.db.WithContext(ctx).Model(&Workflow{}).Where("id = ?", id).Update("status", status).Error
}

// Step operations
func (r *repository) CreateStep(ctx context.Context, step *WorkflowStep) error {
	return r.db.WithContext(ctx).Create(step).Error
}

func (r *repository) UpdateStep(ctx context.Context, step *WorkflowStep) error {
	return r.db.WithContext(ctx).Save(step).Error
}

func (r *repository) DeleteStep(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&WorkflowStep{}, id).Error
}

func (r *repository) GetStepByID(ctx context.Context, id uuid.UUID) (*WorkflowStep, error) {
	var step WorkflowStep
	err := r.db.WithContext(ctx).First(&step, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &step, nil
}

func (r *repository) ListSteps(ctx context.Context, filter *WorkflowStepFilter) ([]WorkflowStep, int64, error) {
	var steps []WorkflowStep
	var total int64

	query := r.db.WithContext(ctx).Model(&WorkflowStep{})

	if filter != nil {
		if filter.WorkflowID != nil {
			query = query.Where("workflow_id = ?", filter.WorkflowID)
		}
		if filter.StepType != nil {
			query = query.Where("step_type = ?", filter.StepType)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.AssignedTo != nil {
			query = query.Where("assigned_to = ?", filter.AssignedTo)
		}
		if filter.StartDate != nil {
			query = query.Where("created_at >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("created_at <= ?", filter.EndDate)
		}
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// Order by step_order for consistent retrieval
	query = query.Order("step_order asc")

	err = query.Find(&steps).Error
	if err != nil {
		return nil, 0, err
	}

	return steps, total, nil
}

func (r *repository) UpdateStepStatus(ctx context.Context, id uuid.UUID, status StepStatus) error {
	return r.db.WithContext(ctx).Model(&WorkflowStep{}).Where("id = ?", id).Update("status", status).Error
}

// Transition operations
func (r *repository) CreateTransition(ctx context.Context, transition *WorkflowTransition) error {
	return r.db.WithContext(ctx).Create(transition).Error
}

func (r *repository) UpdateTransition(ctx context.Context, transition *WorkflowTransition) error {
	return r.db.WithContext(ctx).Save(transition).Error
}

func (r *repository) DeleteTransition(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&WorkflowTransition{}, id).Error
}

func (r *repository) GetTransitionByID(ctx context.Context, id uuid.UUID) (*WorkflowTransition, error) {
	var transition WorkflowTransition
	err := r.db.WithContext(ctx).First(&transition, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &transition, nil
}

func (r *repository) ListTransitions(ctx context.Context, filter *WorkflowTransitionFilter) ([]WorkflowTransition, int64, error) {
	var transitions []WorkflowTransition
	var total int64

	query := r.db.WithContext(ctx).Model(&WorkflowTransition{})

	if filter != nil {
		if filter.FromStepID != nil {
			query = query.Where("from_step_id = ?", filter.FromStepID)
		}
		if filter.ToStepID != nil {
			query = query.Where("to_step_id = ?", filter.ToStepID)
		}
		if filter.OnEvent != nil {
			query = query.Where("on_event = ?", *filter.OnEvent)
		}
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	err = query.Find(&transitions).Error
	if err != nil {
		return nil, 0, err
	}

	return transitions, total, nil
}

// Execution operations
func (r *repository) CreateExecution(ctx context.Context, execution *WorkflowExecution) error {
	return r.db.WithContext(ctx).Create(execution).Error
}

func (r *repository) UpdateExecution(ctx context.Context, execution *WorkflowExecution) error {
	return r.db.WithContext(ctx).Save(execution).Error
}

func (r *repository) GetExecutionByID(ctx context.Context, id uuid.UUID) (*WorkflowExecution, error) {
	var execution WorkflowExecution
	err := r.db.WithContext(ctx).First(&execution, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

func (r *repository) ListExecutions(ctx context.Context, filter *WorkflowExecutionFilter) ([]WorkflowExecution, int64, error) {
	var executions []WorkflowExecution
	var total int64

	query := r.db.WithContext(ctx).Model(&WorkflowExecution{})

	if filter != nil {
		if filter.WorkflowID != nil {
			query = query.Where("workflow_id = ?", filter.WorkflowID)
		}
		if filter.Status != nil {
			query = query.Where("status = ?", filter.Status)
		}
		if filter.StartDate != nil {
			query = query.Where("started_at >= ?", filter.StartDate)
		}
		if filter.EndDate != nil {
			query = query.Where("started_at <= ?", filter.EndDate)
		}
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	if filter != nil && filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query = query.Offset(offset).Limit(filter.PageSize)
	}

	// Order by started_at for most recent executions first
	query = query.Order("started_at desc")

	err = query.Find(&executions).Error
	if err != nil {
		return nil, 0, err
	}

	return executions, total, nil
}

func (r *repository) CancelActiveExecutions(ctx context.Context, workflowID uuid.UUID) error {
	// Marks all active executions as cancelled
	return r.db.WithContext(ctx).Model(&WorkflowExecution{}).
		Where("workflow_id = ? AND status IN ?", workflowID, []WorkflowStatus{WorkflowStatusActive, WorkflowStatusPending}).
		Update("status", WorkflowStatusCancelled).Error
}

// Step execution operations
func (r *repository) CreateStepExecution(ctx context.Context, execution *WorkflowStepExecution) error {
	return r.db.WithContext(ctx).Create(execution).Error
}

func (r *repository) UpdateStepExecution(ctx context.Context, execution *WorkflowStepExecution) error {
	return r.db.WithContext(ctx).Save(execution).Error
}

func (r *repository) GetStepExecutionByID(ctx context.Context, id uuid.UUID) (*WorkflowStepExecution, error) {
	var execution WorkflowStepExecution
	err := r.db.WithContext(ctx).First(&execution, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

func (r *repository) ListStepExecutions(ctx context.Context, executionID uuid.UUID) ([]WorkflowStepExecution, error) {
	var executions []WorkflowStepExecution
	err := r.db.WithContext(ctx).
		Where("execution_id = ?", executionID).
		Order("started_at asc").
		Find(&executions).Error
	if err != nil {
		return nil, err
	}
	return executions, nil
}

// Agent link operations
func (r *repository) CreateAgentLink(ctx context.Context, link *WorkflowAgentLink) error {
	return r.db.WithContext(ctx).Create(link).Error
}

func (r *repository) GetAgentLinksByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]WorkflowAgentLink, error) {
	var links []WorkflowAgentLink
	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Find(&links).Error
	if err != nil {
		return nil, err
	}
	return links, nil
}

// CreateWorkflow creates a new workflow
func (r *repository) CreateWorkflow(ctx context.Context, workflow *Workflow) error {
	r.logger.WithFields(logrus.Fields{
		"name":       workflow.Name,
		"creator_id": workflow.CreatedBy,
	}).Info("Creating new workflow")

	// Ensure proper initialization of maps
	if workflow.Config == nil {
		workflow.Config = datatypes.JSON("{}")
	}
	if workflow.WorkflowMetadata == nil {
		metadata := map[string]interface{}{
			"version":    "1.0.0",
			"created_at": time.Now().UTC(),
			"creator_id": workflow.CreatedBy,
		}
		if jsonData, err := json.Marshal(metadata); err == nil {
			workflow.WorkflowMetadata = datatypes.JSON(jsonData)
		}
	}

	// Initialize tags if nil
	if workflow.Tags == nil {
		workflow.Tags = pq.StringArray{}
	}

	result := r.db.WithContext(ctx).Create(workflow)
	if result.Error != nil {
		r.logger.WithError(result.Error).Error("Failed to create workflow")
		return result.Error
	}

	return nil
}
