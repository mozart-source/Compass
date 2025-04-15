package handlers

import (
	"net/http"
	"strconv"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/project"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ProjectHandler handles HTTP requests for project operations
type ProjectHandler struct {
	service project.Service
}

// NewProjectHandler creates a new ProjectHandler instance
func NewProjectHandler(service project.Service) *ProjectHandler {
	return &ProjectHandler{service: service}
}

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project with the provided information
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param project body dto.CreateProjectRequest true "Project creation request"
// @Success 201 {object} dto.ProjectResponse "Project created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/projects [post]
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req dto.CreateProjectRequest
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

	// Get organization ID from context
	orgID, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization context not found"})
		return
	}

	input := project.CreateProjectInput{
		Name:           req.Name,
		Description:    req.Description,
		Status:         req.Status,
		OrganizationID: orgID.(uuid.UUID),
		CreatorID:      creatorID,
	}

	createdProject, err := h.service.CreateProject(c.Request.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == project.ErrInvalidInput {
			statusCode = http.StatusBadRequest
		} else if err == project.ErrProjectNameExists {
			statusCode = http.StatusConflict
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": dto.ProjectToResponse(createdProject)})
}

// GetProject godoc
// @Summary Get a project by ID
// @Description Get detailed information about a specific project
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID" format(uuid)
// @Success 200 {object} dto.ProjectResponse "Project details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid project ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/projects/{id} [get]
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	// Get organization ID from context
	orgID, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization context not found"})
		return
	}

	// Convert orgID to uuid.UUID
	orgUUID, ok := orgID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid organization ID format"})
		return
	}

	// Get the project
	proj, err := h.service.GetProject(c.Request.Context(), id)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == project.ErrProjectNotFound {
			statuscode = http.StatusNotFound
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	// Verify project belongs to organization
	if proj.OrganizationID != orgUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "project does not belong to the organization"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.ProjectToResponse(proj)})
}

// GetProjectDetails godoc
// @Summary Get detailed project information
// @Description Get project details including members and task counts
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID" format(uuid)
// @Success 200 {object} dto.ProjectDetailsResponse "Project details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid project ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/projects/{id}/details [get]
func (h *ProjectHandler) GetProjectDetails(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	details, err := h.service.GetProjectDetails(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == project.ErrProjectNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	response := dto.ProjectDetailsResponse{
		Project:      *dto.ProjectToResponse(details.Project),
		MembersCount: details.MembersCount,
		TasksCount:   details.TasksCount,
		Members:      make([]dto.MemberResponse, len(details.Members)),
	}

	for i, member := range details.Members {
		response.Members[i] = dto.MemberResponse{
			UserID:   member.UserID,
			Role:     member.Role,
			JoinedAt: member.JoinedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// ListProjects godoc
// @Summary List all projects
// @Description Get a paginated list of projects with optional filters
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 0)"
// @Param pageSize query int false "Number of items per page (default: 10)"
// @Success 200 {object} dto.ProjectListResponse "List of projects retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid pagination parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/projects [get]
func (h *ProjectHandler) ListProjects(c *gin.Context) {
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

	// Get organization ID from context
	orgID, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization context not found"})
		return
	}

	// Convert orgID to uuid.UUID
	orgUUID, ok := orgID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid organization ID format"})
		return
	}

	filter := project.ProjectFilter{
		Page:           page,
		PageSize:       pageSize,
		OrganizationID: &orgUUID,
	}

	// Parse optional filters
	if statusStr := c.Query("status"); statusStr != "" {
		status := project.ProjectStatus(statusStr)
		if status.IsValid() {
			filter.Status = &status
		}
	}
	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}

	projects, total, err := h.service.ListProjects(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert projects to response DTOs
	projectResponses := make([]dto.ProjectResponse, len(projects))
	for i, p := range projects {
		response := dto.ProjectToResponse(&p)
		projectResponses[i] = *response
	}

	response := dto.ProjectListResponse{
		Projects:   projectResponses,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// UpdateProject godoc
// @Summary Update a project
// @Description Update an existing project's information
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID" format(uuid)
// @Param project body dto.UpdateProjectRequest true "Project update information"
// @Success 200 {object} dto.ProjectResponse "Project updated successfully"
// @Failure 400 {object} map[string]string "Invalid request or project ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/projects/{id} [put]
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var req dto.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get organization ID from context
	orgID, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization context not found"})
		return
	}

	// Convert orgID to uuid.UUID
	orgUUID, ok := orgID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid organization ID format"})
		return
	}

	// Get existing project to verify ownership
	existingProj, err := h.service.GetProject(c.Request.Context(), id)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == project.ErrProjectNotFound {
			statuscode = http.StatusNotFound
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	// Verify project belongs to organization
	if existingProj.OrganizationID != orgUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "project does not belong to the organization"})
		return
	}

	input := project.UpdateProjectInput{
		Name:        req.Name,
		Description: req.Description,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	// Convert status if provided
	if req.Status != nil {
		status := project.ProjectStatus(*req.Status)
		if !status.IsValid() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status value"})
			return
		}
		input.Status = &status
	}

	updatedProj, err := h.service.UpdateProject(c.Request.Context(), id, input)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == project.ErrProjectNotFound {
			statuscode = http.StatusNotFound
		} else if err == project.ErrInvalidInput {
			statuscode = http.StatusBadRequest
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.ProjectToResponse(updatedProj)})
}

// DeleteProject godoc
// @Summary Delete a project
// @Description Delete an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID" format(uuid)
// @Success 204 "Project deleted successfully"
// @Failure 400 {object} map[string]string "Invalid project ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/projects/{id} [delete]
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	// Get organization ID from context
	orgID, exists := c.Get("org_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organization context not found"})
		return
	}

	// Convert orgID to uuid.UUID
	orgUUID, ok := orgID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid organization ID format"})
		return
	}

	// Get existing project to verify ownership
	existingProj, err := h.service.GetProject(c.Request.Context(), id)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == project.ErrProjectNotFound {
			statuscode = http.StatusNotFound
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	// Verify project belongs to organization
	if existingProj.OrganizationID != orgUUID {
		c.JSON(http.StatusForbidden, gin.H{"error": "project does not belong to the organization"})
		return
	}

	err = h.service.DeleteProject(c.Request.Context(), id)
	if err != nil {
		statuscode := http.StatusInternalServerError
		if err == project.ErrProjectNotFound {
			statuscode = http.StatusNotFound
		}
		c.JSON(statuscode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// AddProjectMember godoc
// @Summary Add a member to a project
// @Description Add a new member to an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID" format(uuid)
// @Param member body dto.AddMemberRequest true "Member information"
// @Success 201 "Member added successfully"
// @Failure 400 {object} map[string]string "Invalid request or project ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/projects/{id}/members [post]
func (h *ProjectHandler) AddProjectMember(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var req dto.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.AddProjectMember(c.Request.Context(), projectID, req.UserID, req.Role)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == project.ErrProjectNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// RemoveProjectMember godoc
// @Summary Remove a member from a project
// @Description Remove a member from an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID" format(uuid)
// @Param userId path string true "User ID" format(uuid)
// @Success 204 "Member removed successfully"
// @Failure 400 {object} map[string]string "Invalid project ID or user ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Project or member not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/projects/{id}/members/{userId} [delete]
func (h *ProjectHandler) RemoveProjectMember(c *gin.Context) {
	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	err = h.service.RemoveProjectMember(c.Request.Context(), projectID, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == project.ErrProjectNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateProjectStatus godoc
// @Summary Update project status
// @Description Update the status of an existing project
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID" format(uuid)
// @Param status body string true "New project status"
// @Success 200 {object} dto.ProjectResponse "Project status updated successfully"
// @Failure 400 {object} map[string]string "Invalid project ID or status"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Project not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/projects/{id}/status [put]
func (h *ProjectHandler) UpdateProjectStatus(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project ID"})
		return
	}

	var status project.ProjectStatus
	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedProject, err := h.service.UpdateProjectStatus(c.Request.Context(), id, status)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == project.ErrProjectNotFound {
			statusCode = http.StatusNotFound
		} else if err == project.ErrInvalidInput {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": dto.ProjectToResponse(updatedProject)})
}
