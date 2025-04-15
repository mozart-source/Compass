package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/workflow"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
)

// WorkflowHandler handles HTTP requests for workflow operations
type WorkflowHandler struct {
	service workflow.Service
}

// NewWorkflowHandler creates a new WorkflowHandler instance
func NewWorkflowHandler(service workflow.Service) *WorkflowHandler {
	return &WorkflowHandler{service: service}
}

// Helper functions to convert between DTO and domain models
func convertCreateRequestToDomain(req dto.CreateWorkflowRequest) workflow.CreateWorkflowRequest {
	var estimatedDuration *int
	if req.EstimatedDuration != nil {
		intVal := int(*req.EstimatedDuration)
		estimatedDuration = &intVal
	}

	// Convert map to JSON
	var config datatypes.JSON
	if req.Config != nil {
		if jsonData, err := json.Marshal(req.Config); err == nil {
			config = datatypes.JSON(jsonData)
		}
	}

	return workflow.CreateWorkflowRequest{
		Name:              req.Name,
		Description:       req.Description,
		WorkflowType:      workflow.WorkflowType(req.WorkflowType),
		OrganizationID:    req.OrganizationID,
		Config:            config,
		AIEnabled:         req.AIEnabled,
		EstimatedDuration: estimatedDuration,
		Deadline:          req.Deadline,
		Tags:              pq.StringArray(req.Tags),
	}
}

func convertUpdateRequestToDomain(req dto.UpdateWorkflowRequest) workflow.UpdateWorkflowRequest {
	var status *workflow.WorkflowStatus
	if req.Status != nil {
		s := workflow.WorkflowStatus(*req.Status)
		status = &s
	}

	var estimatedDuration *int
	if req.EstimatedDuration != nil {
		intVal := int(*req.EstimatedDuration)
		estimatedDuration = &intVal
	}

	// Convert map to JSON
	var config datatypes.JSON
	if req.Config != nil {
		if jsonData, err := json.Marshal(req.Config); err == nil {
			config = datatypes.JSON(jsonData)
		}
	}

	return workflow.UpdateWorkflowRequest{
		Name:              req.Name,
		Description:       req.Description,
		Status:            status,
		Config:            config,
		AIEnabled:         req.AIEnabled,
		EstimatedDuration: estimatedDuration,
		Deadline:          req.Deadline,
		Tags:              pq.StringArray(req.Tags),
	}
}

// CreateWorkflow godoc
// @Summary Create a new workflow
// @Description Create a new workflow with the provided information
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param workflow body dto.CreateWorkflowRequest true "Workflow creation request"
// @Success 201 {object} dto.WorkflowResponse "Workflow created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Insufficient permissions"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows [post]
func (h *WorkflowHandler) CreateWorkflow(c *gin.Context) {
	var req dto.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get creator ID from context (set by auth middleware)
	creatorID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Convert and validate workflow type
	workflowType := workflow.WorkflowType(req.WorkflowType)
	if !workflowType.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow type"})
		return
	}

	domainReq := convertCreateRequestToDomain(req)
	response, err := h.service.CreateWorkflow(c.Request.Context(), domainReq, creatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

// GetWorkflow godoc
// @Summary Get a workflow by ID
// @Description Get detailed information about a specific workflow
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Success 200 {object} dto.WorkflowResponse "Workflow details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid workflow ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id} [get]
func (h *WorkflowHandler) GetWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	// Get organization ID from header
	orgIDStr := c.GetHeader("X-Organization-ID")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID format"})
		return
	}

	response, err := h.service.GetWorkflow(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if workflow belongs to the organization
	if response.Workflow.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "workflow does not belong to the organization"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// ListWorkflows godoc
// @Summary List all workflows
// @Description Get a paginated list of workflows with optional filters
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Number of items per page (default: 10)"
// @Param organization_id query string false "Filter by organization ID"
// @Param workflow_type query string false "Filter by workflow type"
// @Param status query string false "Filter by status"
// @Param creator_id query string false "Filter by creator ID"
// @Success 200 {object} dto.WorkflowListResponse "List of workflows retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid pagination parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows [get]
func (h *WorkflowHandler) ListWorkflows(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page size"})
		return
	}

	// Get organization ID from header
	orgIDStr := c.GetHeader("X-Organization-ID")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID format"})
		return
	}

	filter := &workflow.WorkflowFilter{
		Page:           page,
		PageSize:       pageSize,
		OrganizationID: &orgID,
	}

	// Parse optional filters
	if workflowTypeStr := c.Query("workflow_type"); workflowTypeStr != "" {
		workflowType := workflow.WorkflowType(workflowTypeStr)
		if workflowType.IsValid() {
			filter.WorkflowType = &workflowType
		}
	}
	if statusStr := c.Query("status"); statusStr != "" {
		status := workflow.WorkflowStatus(statusStr)
		if status.IsValid() {
			filter.Status = &status
		}
	}
	if creatorIDStr := c.Query("creator_id"); creatorIDStr != "" {
		if creatorID, err := uuid.Parse(creatorIDStr); err == nil {
			filter.CreatedBy = &creatorID
		}
	}

	response, err := h.service.ListWorkflows(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateWorkflow godoc
// @Summary Update a workflow
// @Description Update an existing workflow's information
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Param workflow body dto.UpdateWorkflowRequest true "Workflow update information"
// @Success 200 {object} dto.WorkflowResponse "Workflow updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id} [put]
func (h *WorkflowHandler) UpdateWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	var req dto.UpdateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get organization ID from header
	orgIDStr := c.GetHeader("X-Organization-ID")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID format"})
		return
	}

	// Verify workflow belongs to organization
	existingWorkflow, err := h.service.GetWorkflow(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if existingWorkflow.Workflow.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "workflow does not belong to the organization"})
		return
	}

	domainReq := convertUpdateRequestToDomain(req)
	response, err := h.service.UpdateWorkflow(c.Request.Context(), id, domainReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// DeleteWorkflow godoc
// @Summary Delete a workflow
// @Description Delete an existing workflow
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Success 204 "Workflow deleted successfully"
// @Failure 400 {object} map[string]string "Invalid workflow ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id} [delete]
func (h *WorkflowHandler) DeleteWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	// Get organization ID from header
	orgIDStr := c.GetHeader("X-Organization-ID")
	if orgIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
		return
	}

	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID format"})
		return
	}

	// Verify workflow belongs to organization
	existingWorkflow, err := h.service.GetWorkflow(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	if existingWorkflow.Workflow.OrganizationID != orgID {
		c.JSON(http.StatusForbidden, gin.H{"error": "workflow does not belong to the organization"})
		return
	}

	if err := h.service.DeleteWorkflow(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// ExecuteWorkflow godoc
// @Summary Execute a workflow
// @Description Start the execution of a workflow
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Success 200 {object} dto.WorkflowResponse "Workflow execution started successfully"
// @Failure 400 {object} map[string]string "Invalid workflow ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/execute [post]
func (h *WorkflowHandler) ExecuteWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	response, err := h.service.ExecuteWorkflow(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// CancelWorkflowExecution godoc
// @Summary Cancel a workflow execution
// @Description Cancel an active workflow execution
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param executionId path string true "Execution ID" format(uuid)
// @Success 200 "Workflow execution cancelled successfully"
// @Failure 400 {object} map[string]string "Invalid execution ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Execution not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/executions/{executionId}/cancel [post]
func (h *WorkflowHandler) CancelWorkflowExecution(c *gin.Context) {
	executionID, err := uuid.Parse(c.Param("executionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution ID"})
		return
	}

	// Get the execution to get the workflow ID
	execution, err := h.service.GetWorkflowExecution(c.Request.Context(), executionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	workflowID := execution.Execution.WorkflowID

	if err := h.service.CancelWorkflowExecution(c.Request.Context(), workflowID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// AnalyzeWorkflow godoc
// @Summary Analyze a workflow
// @Description Get analysis and metrics for a workflow
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Workflow analysis"
// @Failure 400 {object} map[string]string "Invalid workflow ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/analyze [get]
func (h *WorkflowHandler) AnalyzeWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	analysis, err := h.service.AnalyzeWorkflow(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": analysis})
}

// OptimizeWorkflow godoc
// @Summary Optimize a workflow
// @Description Get optimization recommendations for a workflow
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Success 200 {object} map[string]interface{} "Workflow optimization results"
// @Failure 400 {object} map[string]string "Invalid workflow ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/optimize [post]
func (h *WorkflowHandler) OptimizeWorkflow(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	optimization, err := h.service.OptimizeWorkflow(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": optimization})
}

// CreateWorkflowStep godoc
// @Summary Create a new workflow step
// @Description Create a new step for a specific workflow
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Param step body workflow.CreateWorkflowStepRequest true "Step creation request"
// @Success 201 {object} workflow.WorkflowStepResponse "Step created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Workflow not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/steps [post]
func (h *WorkflowHandler) CreateWorkflowStep(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	var req workflow.CreateWorkflowStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate step type
	if !isValidStepType(string(req.StepType)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step type"})
		return
	}

	response, err := h.service.AddWorkflowStep(c.Request.Context(), id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": response})
}

// ListWorkflowSteps godoc
// @Summary List all steps for a workflow
// @Description Get a list of all steps for a specific workflow
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Success 200 {object} workflow.WorkflowStepListResponse "List of steps retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid workflow ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/steps [get]
func (h *WorkflowHandler) ListWorkflowSteps(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	filter := &workflow.WorkflowStepFilter{
		WorkflowID: &id,
	}

	response, err := h.service.ListWorkflowSteps(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetWorkflowStep godoc
// @Summary Get a workflow step by ID
// @Description Get detailed information about a specific workflow step
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Param stepId path string true "Step ID" format(uuid)
// @Success 200 {object} workflow.WorkflowStepResponse "Step details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Step not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/steps/{stepId} [get]
func (h *WorkflowHandler) GetWorkflowStep(c *gin.Context) {
	_, err := uuid.Parse(c.Param("id")) // Workflow ID
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	stepID, err := uuid.Parse(c.Param("stepId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step ID"})
		return
	}

	response, err := h.service.GetWorkflowStep(c.Request.Context(), stepID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateWorkflowStep godoc
// @Summary Update a workflow step
// @Description Update an existing workflow step
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Param stepId path string true "Step ID" format(uuid)
// @Param step body workflow.UpdateWorkflowStepRequest true "Step update information"
// @Success 200 {object} workflow.WorkflowStepResponse "Step updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Step not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/steps/{stepId} [put]
func (h *WorkflowHandler) UpdateWorkflowStep(c *gin.Context) {
	_, err := uuid.Parse(c.Param("id")) // Workflow ID
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	stepID, err := uuid.Parse(c.Param("stepId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step ID"})
		return
	}

	var req workflow.UpdateWorkflowStepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate step type if provided
	if req.StepType != nil && !isValidStepType(string(*req.StepType)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step type"})
		return
	}

	response, err := h.service.UpdateWorkflowStep(c.Request.Context(), stepID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// DeleteWorkflowStep godoc
// @Summary Delete a workflow step
// @Description Delete an existing workflow step
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Param stepId path string true "Step ID" format(uuid)
// @Success 204 "Step deleted successfully"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Step not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/steps/{stepId} [delete]
func (h *WorkflowHandler) DeleteWorkflowStep(c *gin.Context) {
	_, err := uuid.Parse(c.Param("id")) // Workflow ID
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	stepID, err := uuid.Parse(c.Param("stepId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid step ID"})
		return
	}

	if err := h.service.DeleteWorkflowStep(c.Request.Context(), stepID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// CreateTransition godoc
// @Summary Create a workflow transition
// @Description Create a new transition between workflow steps
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Param transition body workflow.CreateWorkflowTransitionRequest true "Transition creation request"
// @Success 201 {object} workflow.WorkflowTransitionResponse "Transition created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Workflow or steps not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/transitions [post]
func (h *WorkflowHandler) CreateTransition(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	var req workflow.CreateWorkflowTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify from step exists and belongs to the workflow
	fromStep, err := h.service.GetWorkflowStep(c.Request.Context(), req.FromStepID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "from step not found"})
		return
	}
	if fromStep.Step.WorkflowID != workflowID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from step does not belong to this workflow"})
		return
	}

	// Verify to step exists and belongs to the workflow
	toStep, err := h.service.GetWorkflowStep(c.Request.Context(), req.ToStepID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "to step not found"})
		return
	}
	if toStep.Step.WorkflowID != workflowID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "to step does not belong to this workflow"})
		return
	}

	// Create new transition
	transition := &workflow.WorkflowTransition{
		ID:         uuid.New(),
		FromStepID: req.FromStepID,
		ToStepID:   req.ToStepID,
		Conditions: req.Conditions,
		Triggers:   req.Triggers,
		OnEvent:    req.OnEvent,
	}

	if err := h.service.GetRepo().CreateTransition(c.Request.Context(), transition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": &workflow.WorkflowTransitionResponse{Transition: transition}})
}

// GetTransition godoc
// @Summary Get a workflow transition by ID
// @Description Get detailed information about a specific workflow transition
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Param transitionId path string true "Transition ID" format(uuid)
// @Success 200 {object} workflow.WorkflowTransitionResponse "Transition details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Transition not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/transitions/{transitionId} [get]
func (h *WorkflowHandler) GetTransition(c *gin.Context) {
	_, err := uuid.Parse(c.Param("id")) // Workflow ID
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	transitionID, err := uuid.Parse(c.Param("transitionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transition ID"})
		return
	}

	transition, err := h.service.GetRepo().GetTransitionByID(c.Request.Context(), transitionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transition not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": &workflow.WorkflowTransitionResponse{Transition: transition}})
}

// ListTransitions godoc
// @Summary List transitions for a workflow
// @Description Get a list of all transitions for a specific workflow
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Success 200 {object} workflow.WorkflowTransitionListResponse "List of transitions retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid workflow ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/transitions [get]
func (h *WorkflowHandler) ListTransitions(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	// Get all steps for this workflow
	filter := &workflow.WorkflowStepFilter{
		WorkflowID: &workflowID,
	}
	stepsResponse, err := h.service.ListWorkflowSteps(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(stepsResponse.Steps) == 0 {
		c.JSON(http.StatusOK, gin.H{"data": &workflow.WorkflowTransitionListResponse{
			Transitions: []workflow.WorkflowTransition{},
			Total:       0,
		}})
		return
	}

	// Get all step IDs
	var stepIDs []uuid.UUID
	for _, step := range stepsResponse.Steps {
		stepIDs = append(stepIDs, step.ID)
	}

	// Get all transitions for these steps
	var allTransitions []workflow.WorkflowTransition
	var total int64

	for _, stepID := range stepIDs {
		transitionFilter := &workflow.WorkflowTransitionFilter{
			FromStepID: &stepID,
		}
		transitions, count, err := h.service.GetRepo().ListTransitions(c.Request.Context(), transitionFilter)
		if err != nil {
			continue // Skip errors
		}
		allTransitions = append(allTransitions, transitions...)
		total += count
	}

	c.JSON(http.StatusOK, gin.H{"data": &workflow.WorkflowTransitionListResponse{
		Transitions: allTransitions,
		Total:       total,
	}})
}

// UpdateTransition godoc
// @Summary Update a workflow transition
// @Description Update an existing workflow transition
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Param transitionId path string true "Transition ID" format(uuid)
// @Param transition body workflow.UpdateWorkflowTransitionRequest true "Transition update information"
// @Success 200 {object} workflow.WorkflowTransitionResponse "Transition updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Transition not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/transitions/{transitionId} [put]
func (h *WorkflowHandler) UpdateTransition(c *gin.Context) {
	_, err := uuid.Parse(c.Param("id")) // Workflow ID
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	transitionID, err := uuid.Parse(c.Param("transitionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transition ID"})
		return
	}

	var req workflow.UpdateWorkflowTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transition, err := h.service.GetRepo().GetTransitionByID(c.Request.Context(), transitionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transition not found"})
		return
	}

	// Update fields
	if req.Conditions != nil {
		transition.Conditions = req.Conditions
	}
	if req.Triggers != nil {
		transition.Triggers = req.Triggers
	}
	if req.OnEvent != nil {
		transition.OnEvent = *req.OnEvent
	}

	if err := h.service.GetRepo().UpdateTransition(c.Request.Context(), transition); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": &workflow.WorkflowTransitionResponse{Transition: transition}})
}

// DeleteTransition godoc
// @Summary Delete a workflow transition
// @Description Delete an existing workflow transition
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Param transitionId path string true "Transition ID" format(uuid)
// @Success 204 "Transition deleted successfully"
// @Failure 400 {object} map[string]string "Invalid ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Transition not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/transitions/{transitionId} [delete]
func (h *WorkflowHandler) DeleteTransition(c *gin.Context) {
	_, err := uuid.Parse(c.Param("id")) // Workflow ID
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	transitionID, err := uuid.Parse(c.Param("transitionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transition ID"})
		return
	}

	if err := h.service.GetRepo().DeleteTransition(c.Request.Context(), transitionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetWorkflowExecution godoc
// @Summary Get a workflow execution by ID
// @Description Get detailed information about a specific workflow execution
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param executionId path string true "Execution ID" format(uuid)
// @Success 200 {object} workflow.WorkflowExecutionResponse "Execution details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid execution ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Execution not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/executions/{executionId} [get]
func (h *WorkflowHandler) GetWorkflowExecution(c *gin.Context) {
	executionID, err := uuid.Parse(c.Param("executionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution ID"})
		return
	}

	response, err := h.service.GetWorkflowExecution(c.Request.Context(), executionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// ListWorkflowExecutions godoc
// @Summary List executions for a workflow
// @Description Get a list of executions for a specific workflow
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Workflow ID" format(uuid)
// @Param status query string false "Filter by status"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Number of items per page (default: 10)"
// @Success 200 {object} workflow.WorkflowExecutionListResponse "List of executions retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid workflow ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/{id}/executions [get]
func (h *WorkflowHandler) ListWorkflowExecutions(c *gin.Context) {
	workflowID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid workflow ID"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page size"})
		return
	}

	filter := &workflow.WorkflowExecutionFilter{
		WorkflowID: &workflowID,
		Page:       page,
		PageSize:   pageSize,
	}

	// Add status filter if provided
	if statusStr := c.Query("status"); statusStr != "" {
		status := workflow.WorkflowStatus(statusStr)
		if status.IsValid() {
			filter.Status = &status
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
	}

	response, err := h.service.ListWorkflowExecutions(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateStepExecution godoc
// @Summary Update a workflow step execution
// @Description Update the status and result of a workflow step execution
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param executionId path string true "Step Execution ID" format(uuid)
// @Param execution body UpdateStepExecutionRequest true "Step execution update information"
// @Success 200 {object} map[string]string "Step execution updated successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Step execution not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/step-executions/{executionId} [put]
func (h *WorkflowHandler) UpdateStepExecution(c *gin.Context) {
	executionID, err := uuid.Parse(c.Param("executionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution ID"})
		return
	}

	var req UpdateStepExecutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the step execution
	stepExecution, err := h.service.GetRepo().GetStepExecutionByID(c.Request.Context(), executionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "step execution not found"})
		return
	}

	// Update fields
	if req.Status != nil {
		stepExecution.Status = workflow.StepStatus(*req.Status)
	}
	if req.Result != nil {
		stepExecution.Result = req.Result
	}

	// If status is completed or failed, set completed time
	if stepExecution.Status == workflow.StepStatusCompleted ||
		stepExecution.Status == workflow.StepStatusFailed {
		now := time.Now()
		stepExecution.CompletedAt = &now
	}

	// Update the step execution
	if err := h.service.GetRepo().UpdateStepExecution(c.Request.Context(), stepExecution); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get the step to check for auto-advance
	step, err := h.service.GetWorkflowStep(c.Request.Context(), stepExecution.StepID)
	if err == nil && step.Step.AutoAdvance && stepExecution.Status == workflow.StepStatusCompleted {
		// Process next steps if the step is completed and auto-advance is enabled
		// This would typically be handled by the executor, but we'll trigger it manually here
		if h.service.GetExecutor() != nil {
			go func() {
				ctx := context.Background() // Use a new context for async execution
				_ = h.service.GetExecutor().ProcessTransitions(ctx, step.Step, stepExecution, "on_approve")
			}()
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Step execution updated successfully"})
}

// ApproveOrRejectStepRequest represents the request body for approving or rejecting a step
type ApproveOrRejectStepRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ApproveStepExecution godoc
// @Summary Approve a workflow step execution
// @Description Approve a pending workflow step execution
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param executionId path string true "Step Execution ID" format(uuid)
// @Param reason body ApproveOrRejectStepRequest false "Approval reason"
// @Success 200 {object} map[string]string "Step execution approved successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 403 {object} map[string]string "Not authorized or step not approvable"
// @Failure 404 {object} map[string]string "Step execution not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/step-executions/{executionId}/approve [post]
func (h *WorkflowHandler) ApproveStepExecution(c *gin.Context) {
	executionID, err := uuid.Parse(c.Param("executionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution ID"})
		return
	}

	var req ApproveOrRejectStepRequest
	_ = c.ShouldBindJSON(&req) // Ignore error if body is empty

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	if err := h.service.ApproveStepExecution(c.Request.Context(), executionID, userID, req.Reason); err != nil {
		// A more robust error handling can be done here to check specific error types
		if err.Error() == "step execution not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "not authorized" || err.Error() == "step is not of type approval or is not pending" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Step execution approved successfully"})
}

// RejectStepExecution godoc
// @Summary Reject a workflow step execution
// @Description Reject a pending workflow step execution
// @Tags workflows
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param executionId path string true "Step Execution ID" format(uuid)
// @Param reason body ApproveOrRejectStepRequest true "Rejection reason"
// @Success 200 {object} map[string]string "Step execution rejected successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 403 {object} map[string]string "Not authorized or step not approvable"
// @Failure 404 {object} map[string]string "Step execution not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/workflows/step-executions/{executionId}/reject [post]
func (h *WorkflowHandler) RejectStepExecution(c *gin.Context) {
	executionID, err := uuid.Parse(c.Param("executionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution ID"})
		return
	}

	var req ApproveOrRejectStepRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "a reason for rejection is required"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	if err := h.service.RejectStepExecution(c.Request.Context(), executionID, userID, req.Reason); err != nil {
		// A more robust error handling can be done here to check specific error types
		if err.Error() == "step execution not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "not authorized" || err.Error() == "step is not of type approval or is not pending" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Step execution rejected successfully"})
}

// UpdateStepExecutionRequest represents the request body for updating a step execution
type UpdateStepExecutionRequest struct {
	Status *string        `json:"status,omitempty"`
	Result datatypes.JSON `json:"result,omitempty"`
}

// Helper function to check if a step type is valid
func isValidStepType(stepType string) bool {
	validTypes := map[string]bool{
		string(workflow.StepTypeManual):       true,
		string(workflow.StepTypeAutomated):    true,
		string(workflow.StepTypeApproval):     true,
		string(workflow.StepTypeNotification): true,
		string(workflow.StepTypeIntegration):  true,
		string(workflow.StepTypeDecision):     true,
		string(workflow.StepTypeAITask):       true,
	}
	return validTypes[stepType]
}
