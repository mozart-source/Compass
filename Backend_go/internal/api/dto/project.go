package dto

import (
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/project"
	"github.com/google/uuid"
)

// CreateProjectRequest represents the request body for creating a new project
// @Description Request body for creating a new project in the system
type CreateProjectRequest struct {
	Name           string                `json:"name" binding:"required" example:"E-commerce Platform"`
	Description    string                `json:"description" example:"A modern e-commerce platform with microservices architecture"`
	Status         project.ProjectStatus `json:"status" example:"Active"`
	OrganizationID uuid.UUID             `json:"organization_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	CreatorID      uuid.UUID             `json:"creator_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	OwnerID        uuid.UUID             `json:"owner_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440002"`
	StartDate      time.Time             `json:"start_date" binding:"required" example:"2024-01-01T00:00:00Z"`
	EndDate        *time.Time            `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`
}

// UpdateProjectRequest represents the request body for updating an existing project
// @Description Request body for updating project information
type UpdateProjectRequest struct {
	Name        *string                `json:"name,omitempty" example:"Updated Project Name"`
	Description *string                `json:"description,omitempty" example:"Updated project description"`
	Status      *project.ProjectStatus `json:"status,omitempty" example:"active"`
	OwnerID     *uuid.UUID             `json:"owner_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440003"`
	StartDate   *time.Time             `json:"start_date,omitempty" example:"2024-01-01T00:00:00Z"`
	EndDate     *time.Time             `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`
}

// ProjectResponse represents a project in API responses
// @Description Detailed project information returned in API responses
type ProjectResponse struct {
	ID             uuid.UUID             `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name           string                `json:"name" example:"Project Alpha"`
	Description    string                `json:"description" example:"A project for managing tasks"`
	Status         project.ProjectStatus `json:"status" example:"active"`
	OrganizationID uuid.UUID             `json:"organization_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	CreatorID      uuid.UUID             `json:"creator_id" example:"550e8400-e29b-41d4-a716-446655440002"`
	OwnerID        uuid.UUID             `json:"owner_id" example:"550e8400-e29b-41d4-a716-446655440003"`
	StartDate      time.Time             `json:"start_date" example:"2024-01-01T00:00:00Z"`
	EndDate        *time.Time            `json:"end_date,omitempty" example:"2024-12-31T23:59:59Z"`
	CreatedAt      time.Time             `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt      time.Time             `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// ProjectDetailsResponse represents detailed project information including members and tasks
// @Description Detailed project information with related data
type ProjectDetailsResponse struct {
	Project      ProjectResponse  `json:"project"`
	MembersCount int64            `json:"members_count" example:"5"`
	TasksCount   int64            `json:"tasks_count" example:"20"`
	Members      []MemberResponse `json:"members"`
}

// MemberResponse represents a project member in API responses
// @Description Project member information
type MemberResponse struct {
	UserID   uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role     string    `json:"role" example:"Developer"`
	JoinedAt time.Time `json:"joined_at" example:"2024-03-15T09:00:00Z"`
}

// ProjectListResponse represents a paginated list of projects
// @Description Paginated list of projects with metadata
type ProjectListResponse struct {
	Projects   []ProjectResponse `json:"projects"`
	TotalCount int64             `json:"total_count" example:"100"`
	Page       int               `json:"page" example:"1"`
	PageSize   int               `json:"page_size" example:"20"`
}

// AddMemberRequest represents the request body for adding a member to a project
// @Description Request body for adding a new member to a project
type AddMemberRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role   string    `json:"role" binding:"required" example:"Developer"`
}

// Convert domain Project to ProjectResponse
func ProjectToResponse(p *project.Project) *ProjectResponse {
	return &ProjectResponse{
		ID:             p.ID,
		Name:           p.Name,
		Description:    p.Description,
		Status:         p.Status,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
		CreatorID:      p.CreatorID,
		OrganizationID: p.OrganizationID,
		OwnerID:        p.OwnerID,
		StartDate:      p.StartDate,
		EndDate:        p.EndDate,
	}
}

// Convert domain Projects to ProjectResponses
func ProjectsToResponse(projects []project.Project) []*ProjectResponse {
	response := make([]*ProjectResponse, len(projects))
	for i, p := range projects {
		response[i] = ProjectToResponse(&p)
	}
	return response
}
