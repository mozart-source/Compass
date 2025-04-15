package routes

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// ProjectRoutes handles the setup of project-related routes
type ProjectRoutes struct {
	handler   *handlers.ProjectHandler
	jwtSecret string
}

// NewProjectRoutes creates a new ProjectRoutes instance
func NewProjectRoutes(handler *handlers.ProjectHandler, jwtSecret string) *ProjectRoutes {
	return &ProjectRoutes{
		handler:   handler,
		jwtSecret: jwtSecret,
	}
}

// RegisterRoutes registers all project-related routes
func (pr *ProjectRoutes) RegisterRoutes(router *gin.Engine, cache *middleware.CacheMiddleware) {
	// Create a project group with authentication middleware
	projectGroup := router.Group("/api/projects")
	projectGroup.Use(middleware.NewAuthMiddleware(pr.jwtSecret))

	// @Summary Create a new project
	// @Description Create a new project with the provided information
	// @Tags projects
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param project body dto.CreateProjectRequest true "Project creation information"
	// @Success 201 {object} dto.ProjectResponse "Project created successfully"
	// @Failure 400 {object} map[string]string "Invalid request"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Insufficient permissions"
	// @Failure 409 {object} map[string]string "Project name already exists"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/projects [post]
	projectGroup.POST("", cache.CacheInvalidate("projects:*"), pr.handler.CreateProject)

	// @Summary Get all projects
	// @Description Get all projects with pagination and filtering
	// @Tags projects
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param page query int false "Page number (default: 0)"
	// @Param pageSize query int false "Page size (default: 10)"
	// @Param status query string false "Filter by status (Active, Completed, Archived, On Hold)"
	// @Param name query string false "Filter by project name"
	// @Success 200 {object} dto.ProjectListResponse "List of projects"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Insufficient permissions"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/projects [get]
	projectGroup.GET("", cache.CacheResponse(), pr.handler.ListProjects)

	// @Summary Get a project by ID
	// @Description Get detailed information about a specific project
	// @Tags projects
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param id path string true "Project ID" format(uuid)
	// @Success 200 {object} dto.ProjectResponse "Project details"
	// @Failure 400 {object} map[string]string "Invalid project ID"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Insufficient permissions"
	// @Failure 404 {object} map[string]string "Project not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/projects/{id} [get]
	projectGroup.GET("/:id", cache.CacheResponse(), pr.handler.GetProject)

	// @Summary Get detailed project information
	// @Description Get detailed project information including members and task counts
	// @Tags projects
	// @Accept json
	// @Produce json
	// @Security BearerAuth
	// @Param id path string true "Project ID" format(uuid)
	// @Success 200 {object} dto.ProjectDetailsResponse "Project details with members"
	// @Failure 400 {object} map[string]string "Invalid project ID"
	// @Failure 401 {object} map[string]string "Unauthorized"
	// @Failure 403 {object} map[string]string "Insufficient permissions"
	// @Failure 404 {object} map[string]string "Project not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/projects/{id}/details [get]
	projectGroup.GET("/:id/details", cache.CacheResponse(), pr.handler.GetProjectDetails)

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
	// @Failure 403 {object} map[string]string "Insufficient permissions"
	// @Failure 404 {object} map[string]string "Project not found"
	// @Failure 409 {object} map[string]string "Project name already exists"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/projects/{id} [put]
	projectGroup.PUT("/:id", cache.CacheInvalidate("projects:*"), pr.handler.UpdateProject)

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
	// @Failure 403 {object} map[string]string "Insufficient permissions"
	// @Failure 404 {object} map[string]string "Project not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/projects/{id} [delete]
	projectGroup.DELETE("/:id", cache.CacheInvalidate("projects:*"), pr.handler.DeleteProject)

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
	// @Failure 403 {object} map[string]string "Insufficient permissions"
	// @Failure 404 {object} map[string]string "Project not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/projects/{id}/members [post]
	projectGroup.POST("/:id/members", cache.CacheInvalidate("projects:*"), pr.handler.AddProjectMember)

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
	// @Failure 403 {object} map[string]string "Insufficient permissions"
	// @Failure 404 {object} map[string]string "Project or member not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/projects/{id}/members/{userId} [delete]
	projectGroup.DELETE("/:id/members/:userId", cache.CacheInvalidate("projects:*"), pr.handler.RemoveProjectMember)

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
	// @Failure 403 {object} map[string]string "Insufficient permissions"
	// @Failure 404 {object} map[string]string "Project not found"
	// @Failure 500 {object} map[string]string "Internal server error"
	// @Router /api/projects/{id}/status [put]
	projectGroup.PUT("/:id/status", cache.CacheInvalidate("projects:*"), pr.handler.UpdateProjectStatus)
}
