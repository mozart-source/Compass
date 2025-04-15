package dto

import (
	"encoding/json"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/workflow"
	"github.com/google/uuid"
)

// CreateWorkflowRequest represents the request to create a new workflow
type CreateWorkflowRequest struct {
	Name              string                 `json:"name" binding:"required"`
	Description       string                 `json:"description"`
	WorkflowType      string                 `json:"workflow_type" binding:"required"`
	OrganizationID    uuid.UUID              `json:"organization_id" binding:"required"`
	Config            map[string]interface{} `json:"config"`
	AIEnabled         bool                   `json:"ai_enabled"`
	Tags              []string               `json:"tags"`
	EstimatedDuration *int64                 `json:"estimated_duration"`
	Deadline          *time.Time             `json:"deadline"`
}

// UpdateWorkflowRequest represents the request to update a workflow
type UpdateWorkflowRequest struct {
	Name              *string                `json:"name"`
	Description       *string                `json:"description"`
	Status            *string                `json:"status"`
	Config            map[string]interface{} `json:"config"`
	AIEnabled         *bool                  `json:"ai_enabled"`
	Tags              []string               `json:"tags"`
	EstimatedDuration *int64                 `json:"estimated_duration"`
	Deadline          *time.Time             `json:"deadline"`
}

// CreateWorkflowStepRequest represents the request to create a new workflow step
type CreateWorkflowStepRequest struct {
	Name             string                 `json:"name" binding:"required"`
	Description      string                 `json:"description"`
	StepType         string                 `json:"step_type" binding:"required"`
	StepOrder        int                    `json:"step_order" binding:"required"`
	Config           map[string]interface{} `json:"config"`
	Conditions       map[string]interface{} `json:"conditions"`
	Timeout          *int64                 `json:"timeout"`
	IsRequired       *bool                  `json:"is_required"`
	AssignedTo       *uuid.UUID             `json:"assigned_to"`
	AssignedToRoleID *uuid.UUID             `json:"assigned_to_role_id"`
}

// UpdateWorkflowStepRequest represents the request to update a workflow step
type UpdateWorkflowStepRequest struct {
	Name             *string                `json:"name"`
	Description      *string                `json:"description"`
	StepType         *string                `json:"step_type"`
	StepOrder        *int                   `json:"step_order"`
	Status           *string                `json:"status"`
	Config           map[string]interface{} `json:"config"`
	Conditions       map[string]interface{} `json:"conditions"`
	Timeout          *int64                 `json:"timeout"`
	IsRequired       *bool                  `json:"is_required"`
	AssignedTo       *uuid.UUID             `json:"assigned_to"`
	AssignedToRoleID *uuid.UUID             `json:"assigned_to_role_id"`
}

// WorkflowResponse represents the response for a workflow
type WorkflowResponse struct {
	ID                uuid.UUID              `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	WorkflowType      string                 `json:"workflow_type"`
	Status            string                 `json:"status"`
	CreatedBy         uuid.UUID              `json:"created_by"`
	OrganizationID    uuid.UUID              `json:"organization_id"`
	Config            map[string]interface{} `json:"config"`
	AIEnabled         bool                   `json:"ai_enabled"`
	Tags              []string               `json:"tags"`
	EstimatedDuration *int64                 `json:"estimated_duration,omitempty"`
	ActualDuration    *int64                 `json:"actual_duration,omitempty"`
	Deadline          *time.Time             `json:"deadline,omitempty"`
	LastExecutedAt    *time.Time             `json:"last_executed_at,omitempty"`
	SuccessRate       float64                `json:"success_rate"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

// WorkflowListResponse represents the response for a list of workflows
type WorkflowListResponse struct {
	Workflows        []WorkflowResponse `json:"workflows"`
	Total            int64              `json:"total"`
	Page             int                `json:"page"`
	PageSize         int                `json:"page_size"`
	Timeout          *int64             `json:"timeout,omitempty"`
	IsRequired       bool               `json:"is_required"`
	AssignedTo       *uuid.UUID         `json:"assigned_to,omitempty"`
	AssignedToRoleID *uuid.UUID         `json:"assigned_to_role_id,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
}

// WorkflowStepResponse represents the response for a workflow step
type WorkflowStepResponse struct {
	ID               uuid.UUID              `json:"id"`
	WorkflowID       uuid.UUID              `json:"workflow_id"`
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	StepType         string                 `json:"step_type"`
	StepOrder        int                    `json:"step_order"`
	Status           string                 `json:"status"`
	Config           map[string]interface{} `json:"config"`
	Conditions       map[string]interface{} `json:"conditions"`
	Timeout          *int64                 `json:"timeout,omitempty"`
	IsRequired       bool                   `json:"is_required"`
	AssignedTo       *uuid.UUID             `json:"assigned_to,omitempty"`
	AssignedToRoleID *uuid.UUID             `json:"assigned_to_role_id"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// WorkflowStepListResponse represents the response for a list of workflow steps
type WorkflowStepListResponse struct {
	Steps    []WorkflowStepResponse `json:"steps"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

// Helper functions to convert between domain models and DTOs

// WorkflowToResponse converts a workflow domain model to a response DTO
func WorkflowToResponse(w *workflow.Workflow) *WorkflowResponse {
	if w == nil {
		return nil
	}

	var config map[string]interface{}
	if len(w.Config) > 0 {
		_ = json.Unmarshal(w.Config, &config)
	}

	var estimatedDuration, actualDuration *int64
	if w.EstimatedDuration != nil {
		i64 := int64(*w.EstimatedDuration)
		estimatedDuration = &i64
	}
	if w.ActualDuration != nil {
		i64 := int64(*w.ActualDuration)
		actualDuration = &i64
	}

	return &WorkflowResponse{
		ID:                w.ID,
		Name:              w.Name,
		Description:       w.Description,
		WorkflowType:      string(w.WorkflowType),
		Status:            string(w.Status),
		CreatedBy:         w.CreatedBy,
		OrganizationID:    w.OrganizationID,
		Config:            config,
		AIEnabled:         w.AIEnabled,
		Tags:              w.Tags,
		EstimatedDuration: estimatedDuration,
		ActualDuration:    actualDuration,
		Deadline:          w.Deadline,
		LastExecutedAt:    w.LastExecutedAt,
		SuccessRate:       w.SuccessRate,
		CreatedAt:         w.CreatedAt,
		UpdatedAt:         w.UpdatedAt,
	}
}

// WorkflowStepToResponse converts a workflow step domain model to a response DTO
func WorkflowStepToResponse(s *workflow.WorkflowStep) *WorkflowStepResponse {
	if s == nil {
		return nil
	}

	var timeout *int64
	if s.Timeout != nil {
		i64 := int64(*s.Timeout)
		timeout = &i64
	}

	// Convert JSON fields to maps
	var config, conditions map[string]interface{}
	if len(s.Config) > 0 {
		_ = json.Unmarshal(s.Config, &config)
	}
	if len(s.Conditions) > 0 {
		_ = json.Unmarshal(s.Conditions, &conditions)
	}

	return &WorkflowStepResponse{
		ID:               s.ID,
		WorkflowID:       s.WorkflowID,
		Name:             s.Name,
		Description:      s.Description,
		StepType:         string(s.StepType),
		StepOrder:        s.StepOrder,
		Status:           string(s.Status),
		Config:           config,
		Conditions:       conditions,
		Timeout:          timeout,
		IsRequired:       s.IsRequired,
		AssignedTo:       s.AssignedTo,
		AssignedToRoleID: s.AssignedToRoleID,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}
