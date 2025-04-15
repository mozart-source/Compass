package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/task"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TaskHandler handles HTTP requests for task operations
type TaskHandler struct {
	service task.Service
}

// NewTaskHandler creates a new TaskHandler instance
func NewTaskHandler(service task.Service) *TaskHandler {
	return &TaskHandler{service: service}
}

// CreateTask godoc
// @Summary Create a new task
// @Description Create a new task with the provided information
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param task body dto.CreateTaskRequest true "Task creation request"
// @Success 201 {object} dto.TaskResponse "Task created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Insufficient permissions"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req dto.CreateTaskRequest
	validatedModel, exists := c.Get("validated_model")

	if exists {
		// If validation middleware provided the model, use it
		if validatedPtr, ok := validatedModel.(*dto.CreateTaskRequest); ok {
			req = *validatedPtr
		} else {
			log.Errorf("Invalid model type: %T, expected *dto.CreateTaskRequest", validatedModel)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model type from validation"})
			return
		}
	} else {
		// If validation middleware didn't run, do manual binding
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Get creator ID from context (set by auth middleware)
	creatorID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Convert and validate status
	status := task.TaskStatus(req.Status)
	if !status.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	// Convert and validate priority
	priority := task.TaskPriority(req.Priority)
	if !priority.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid priority value"})
		return
	}

	input := task.CreateTaskInput{
		Title:          req.Title,
		Description:    req.Description,
		Status:         status,
		Priority:       priority,
		ProjectID:      req.ProjectID,
		OrganizationID: req.OrganizationID,
		AssigneeID:     req.AssigneeID,
		ReviewerID:     req.ReviewerID,
		CategoryID:     req.CategoryID,
		ParentTaskID:   req.ParentTaskID,
		EstimatedHours: req.EstimatedHours,
		StartDate:      req.StartDate,
		Duration:       req.Duration,
		DueDate:        req.DueDate,
		Dependencies:   req.Dependencies,
		CreatorID:      creatorID,
	}

	createdTask, err := h.service.CreateTask(c.Request.Context(), input)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == task.ErrInvalidInput {
			statuscode = http.StatusBadRequest
		} else if err == task.ErrInvalidCreator {
			statuscode = http.StatusForbidden
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": TaskToResponse(createdTask)})
}

// GetTask godoc
// @Summary Get a task by ID
// @Description Get detailed information about a specific task
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID" format(uuid)
// @Success 200 {object} dto.TaskResponse "Task details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid task ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	tsk, err := h.service.GetTask(c.Request.Context(), id)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == task.ErrTaskNotFound {
			statuscode = http.StatusNotFound
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TaskToResponse(tsk)})
}

// ListTasks godoc
// @Summary List all tasks
// @Description Get a paginated list of tasks with optional filters
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 0)"
// @Param pageSize query int false "Number of items per page (default: 10)"
// @Param organization_id query string false "Filter by organization ID"
// @Param project_id query string false "Filter by project ID"
// @Param status query string false "Filter by status"
// @Param priority query string false "Filter by priority"
// @Param assignee_id query string false "Filter by assignee ID"
// @Param creator_id query string false "Filter by creator ID"
// @Param reviewer_id query string false "Filter by reviewer ID"
// @Success 200 {object} dto.TaskListResponse "List of tasks retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid pagination parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Insufficient permissions"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "0")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

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

	filter := task.TaskFilter{
		Page:     page,
		PageSize: pageSize,
	}

	// Parse optional filters
	if projectIDStr := c.Query("project_id"); projectIDStr != "" {
		if projectID, err := uuid.Parse(projectIDStr); err == nil {
			filter.ProjectID = &projectID
		}
	}
	if statusStr := c.Query("status"); statusStr != "" {
		status := task.TaskStatus(statusStr)
		if status.IsValid() {
			filter.Status = &status
		}
	}
	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority := task.TaskPriority(priorityStr)
		if priority.IsValid() {
			filter.Priority = &priority
		}
	}
	if assigneeIDStr := c.Query("assignee_id"); assigneeIDStr != "" {
		if assigneeID, err := uuid.Parse(assigneeIDStr); err == nil {
			filter.AssigneeID = &assigneeID
		}
	}
	if creatorIDStr := c.Query("creator_id"); creatorIDStr != "" {
		if creatorID, err := uuid.Parse(creatorIDStr); err == nil {
			filter.CreatorID = &creatorID
		}
	}
	if reviewerIDStr := c.Query("reviewer_id"); reviewerIDStr != "" {
		if reviewerID, err := uuid.Parse(reviewerIDStr); err == nil {
			filter.ReviewerID = &reviewerID
		}
	}

	tasks, total, err := h.service.ListTasks(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert tasks to response DTOs
	taskResponses := make([]dto.TaskResponse, len(tasks))
	for i, t := range tasks {
		response := TaskToResponse(&t)
		taskResponses[i] = *response
	}

	response := dto.TaskListResponse{
		Tasks:      taskResponses,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateTask godoc
// @Summary Update a task
// @Description Update an existing task's information
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID" format(uuid)
// @Param task body dto.UpdateTaskRequest true "Task update information"
// @Success 200 {object} dto.TaskResponse "Task updated successfully"
// @Failure 400 {object} map[string]string "Invalid request or task ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var req dto.UpdateTaskRequest
	validatedModel, exists := c.Get("validated_model")

	if exists {
		// If validation middleware provided the model, use it
		if validatedPtr, ok := validatedModel.(*dto.UpdateTaskRequest); ok {
			req = *validatedPtr
		} else {
			log.Errorf("Invalid model type: %T, expected *dto.UpdateTaskRequest", validatedModel)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model type from validation"})
			return
		}
	} else {
		// If validation middleware didn't run, do manual binding
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	input := task.UpdateTaskInput{
		Title:          req.Title,
		Description:    req.Description,
		AssigneeID:     req.AssigneeID,
		ReviewerID:     req.ReviewerID,
		CategoryID:     req.CategoryID,
		EstimatedHours: req.EstimatedHours,
		StartDate:      req.StartDate,
		Duration:       req.Duration,
		DueDate:        req.DueDate,
		Dependencies:   req.Dependencies,
	}

	// Convert status if provided
	if req.Status != nil {
		status := task.TaskStatus(*req.Status)
		if !status.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
			return
		}
		input.Status = &status
	}

	// Convert priority if provided
	if req.Priority != nil {
		priority := task.TaskPriority(*req.Priority)
		if !priority.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid priority value"})
			return
		}
		input.Priority = &priority
	}

	updatedTask, err := h.service.UpdateTask(c.Request.Context(), id, input)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == task.ErrTaskNotFound {
			statuscode = http.StatusNotFound
		} else if err == task.ErrInvalidInput {
			statuscode = http.StatusBadRequest
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TaskToResponse(updatedTask)})
}

// DeleteTask godoc
// @Summary Delete a task
// @Description Delete an existing task
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID" format(uuid)
// @Success 204 "Task deleted successfully"
// @Failure 400 {object} map[string]string "Invalid task ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	err = h.service.DeleteTask(c.Request.Context(), id)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == task.ErrTaskNotFound {
			statuscode = http.StatusNotFound
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetProjectTasks godoc
// @Summary Get tasks for a project
// @Description Get tasks for a specific project
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project_id path string true "Project ID" format(uuid)
// @Param page query int false "Page number (default: 0)"
// @Param pageSize query int false "Number of items per page (default: 10)"
// @Success 200 {object} dto.TaskListResponse "List of tasks retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid project ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Insufficient permissions"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks/project/{project_id} [get]
func (h *TaskHandler) GetProjectTasks(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	pageStr := c.DefaultQuery("page", "0")
	pageSizeStr := c.DefaultQuery("pageSize", "10")

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

	filter := task.TaskFilter{
		Page:     page,
		PageSize: pageSize,
	}

	tasks, total, err := h.service.GetProjectTasks(c.Request.Context(), projectID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert tasks to response DTOs
	taskResponses := make([]dto.TaskResponse, len(tasks))
	for i, t := range tasks {
		response := TaskToResponse(&t)
		taskResponses[i] = *response
	}

	response := dto.TaskListResponse{
		Tasks:      taskResponses,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateTaskStatus godoc
// @Summary Update task status
// @Description Update the status of a task
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID" format(uuid)
// @Param status body dto.UpdateTaskStatusRequest true "Task status update information"
// @Success 200 {object} dto.TaskResponse "Task status updated successfully"
// @Failure 400 {object} map[string]string "Invalid request or task ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Insufficient permissions"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks/{id}/status [patch]
func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var req dto.UpdateTaskStatusRequest
	validatedModel, exists := c.Get("validated_model")

	if exists {
		// If validation middleware provided the model, use it
		if validatedPtr, ok := validatedModel.(*dto.UpdateTaskStatusRequest); ok {
			req = *validatedPtr
		} else {
			log.Errorf("Invalid model type: %T, expected *dto.UpdateTaskStatusRequest", validatedModel)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model type from validation"})
			return
		}
	} else {
		// If validation middleware didn't run, do manual binding
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	status := task.TaskStatus(req.Status)
	if !status.IsValid() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
		return
	}

	updatedTask, err := h.service.UpdateTaskStatus(c.Request.Context(), id, status)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == task.ErrTaskNotFound {
			statuscode = http.StatusNotFound
		} else if err == task.ErrInvalidInput || err == task.ErrInvalidTransition {
			statuscode = http.StatusBadRequest
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TaskToResponse(updatedTask)})
}

// AssignTask godoc
// @Summary Assign a task
// @Description Assign a task to a user
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID" format(uuid)
// @Param assign body dto.AssignTaskRequest true "Task assignment information"
// @Success 200 {object} dto.TaskResponse "Task assigned successfully"
// @Failure 400 {object} map[string]string "Invalid request or task ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Insufficient permissions"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks/{id}/assign [patch]
func (h *TaskHandler) AssignTask(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	var req dto.AssignTaskRequest
	validatedModel, exists := c.Get("validated_model")

	if exists {
		// If validation middleware provided the model, use it
		if validatedPtr, ok := validatedModel.(*dto.AssignTaskRequest); ok {
			req = *validatedPtr
		} else {
			log.Errorf("Invalid model type: %T, expected *dto.AssignTaskRequest", validatedModel)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model type from validation"})
			return
		}
	} else {
		// If validation middleware didn't run, do manual binding
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	assigneeID, err := uuid.Parse(req.AssigneeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignee ID"})
		return
	}

	updatedTask, err := h.service.AssignTask(c.Request.Context(), id, assigneeID)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == task.ErrTaskNotFound {
			statuscode = http.StatusNotFound
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": TaskToResponse(updatedTask)})
}

// GetTaskAnalytics godoc
// @Summary Get task analytics
// @Description Get analytics data for a specific task
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID" format(uuid)
// @Param start_time query string true "Start time (RFC3339)" format(date-time)
// @Param end_time query string true "End time (RFC3339)" format(date-time)
// @Param page query int false "Page number (default: 0)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} dto.TaskAnalyticsListResponse "Task analytics retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks/{id}/analytics [get]
func (h *TaskHandler) GetTaskAnalytics(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	// Parse time strings to time.Time
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, expected RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, expected RFC3339"})
		return
	}

	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "0")
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

	// Get analytics data
	analytics, total, err := h.service.GetTaskAnalytics(
		c.Request.Context(),
		taskID,
		startTime,
		endTime,
		page,
		pageSize,
	)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == task.ErrTaskNotFound {
			statuscode = http.StatusNotFound
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	// Convert domain entities to DTO responses
	responseItems := make([]dto.TaskAnalyticsResponse, len(analytics))
	for i, item := range analytics {
		// Parse metadata JSON
		var metadata map[string]interface{}
		if item.Metadata != "" {
			if err := json.Unmarshal([]byte(item.Metadata), &metadata); err != nil {
				// Just use empty metadata if parsing fails
				metadata = make(map[string]interface{})
			}
		} else {
			metadata = make(map[string]interface{})
		}

		responseItems[i] = dto.TaskAnalyticsResponse{
			ID:        item.ID,
			TaskID:    item.TaskID,
			UserID:    item.UserID,
			Action:    item.Action,
			Timestamp: item.Timestamp,
			Metadata:  metadata,
		}
	}

	response := dto.TaskAnalyticsListResponse{
		Analytics:  responseItems,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetUserTaskAnalytics godoc
// @Summary Get user task analytics
// @Description Get analytics data for tasks associated with a user
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_time query string true "Start time (RFC3339)" format(date-time)
// @Param end_time query string true "End time (RFC3339)" format(date-time)
// @Param page query int false "Page number (default: 0)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} dto.TaskAnalyticsListResponse "User task analytics retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks/analytics/user [get]
func (h *TaskHandler) GetUserTaskAnalytics(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Parse time strings to time.Time
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, expected RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, expected RFC3339"})
		return
	}

	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "0")
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

	// Get analytics data
	analytics, total, err := h.service.GetUserTaskAnalytics(
		c.Request.Context(),
		userID,
		startTime,
		endTime,
		page,
		pageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert domain entities to DTO responses
	responseItems := make([]dto.TaskAnalyticsResponse, len(analytics))
	for i, item := range analytics {
		// Parse metadata JSON
		var metadata map[string]interface{}
		if item.Metadata != "" {
			if err := json.Unmarshal([]byte(item.Metadata), &metadata); err != nil {
				// Just use empty metadata if parsing fails
				metadata = make(map[string]interface{})
			}
		} else {
			metadata = make(map[string]interface{})
		}

		responseItems[i] = dto.TaskAnalyticsResponse{
			ID:        item.ID,
			TaskID:    item.TaskID,
			UserID:    item.UserID,
			Action:    item.Action,
			Timestamp: item.Timestamp,
			Metadata:  metadata,
		}
	}

	response := dto.TaskAnalyticsListResponse{
		Analytics:  responseItems,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetTaskActivitySummary godoc
// @Summary Get task activity summary
// @Description Get a summary of activity counts by action type for a task
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID" format(uuid)
// @Param start_time query string true "Start time (RFC3339)" format(date-time)
// @Param end_time query string true "End time (RFC3339)" format(date-time)
// @Success 200 {object} map[string]interface{} "Task activity summary retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks/{id}/analytics/summary [get]
func (h *TaskHandler) GetTaskActivitySummary(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	// Parse time strings to time.Time
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, expected RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, expected RFC3339"})
		return
	}

	// Get summary data
	summary, err := h.service.GetTaskActivitySummary(
		c.Request.Context(),
		taskID,
		startTime,
		endTime,
	)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == task.ErrTaskNotFound {
			statuscode = http.StatusNotFound
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"task_id":       summary.TaskID,
		"action_counts": summary.ActionCounts,
		"start_time":    summary.StartTime,
		"end_time":      summary.EndTime,
		"total_actions": summary.TotalActions,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetUserTaskActivitySummary godoc
// @Summary Get user task activity summary
// @Description Get a summary of activity counts by action type for a user's tasks
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_time query string true "Start time (RFC3339)" format(date-time)
// @Param end_time query string true "End time (RFC3339)" format(date-time)
// @Success 200 {object} map[string]interface{} "User task activity summary retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks/analytics/user/summary [get]
func (h *TaskHandler) GetUserTaskActivitySummary(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Parse time strings to time.Time
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, expected RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, expected RFC3339"})
		return
	}

	// Get summary data
	summary, err := h.service.GetUserTaskActivitySummary(
		c.Request.Context(),
		userID,
		startTime,
		endTime,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"user_id":       summary.UserID,
		"action_counts": summary.ActionCounts,
		"start_time":    summary.StartTime,
		"end_time":      summary.EndTime,
		"total_actions": summary.TotalActions,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// RecordTaskActivity godoc
// @Summary Record task activity
// @Description Manually record a task activity for analytics
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Task ID" format(uuid)
// @Param activity body dto.RecordUserActivityRequest true "Activity details"
// @Success 201 "Activity recorded successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Task not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/tasks/{id}/analytics/record [post]
func (h *TaskHandler) RecordTaskActivity(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var request dto.RecordUserActivityRequest
	validatedModel, exists := c.Get("validated_model")

	if exists {
		// If validation middleware provided the model, use it
		if validatedPtr, ok := validatedModel.(*dto.RecordUserActivityRequest); ok {
			request = *validatedPtr
		} else {
			log.Errorf("Invalid model type: %T, expected *dto.RecordUserActivityRequest", validatedModel)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model type from validation"})
			return
		}
	} else {
		// If validation middleware didn't run, do manual binding
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Verify task exists
	_, err = h.service.GetTask(c.Request.Context(), taskID)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == task.ErrTaskNotFound {
			statuscode = http.StatusNotFound
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	input := task.RecordTaskActivityInput{
		TaskID:    taskID,
		UserID:    userID,
		Action:    request.Action,
		Metadata:  request.Metadata,
		Timestamp: time.Now(),
	}

	if err := h.service.RecordTaskActivity(c.Request.Context(), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}
