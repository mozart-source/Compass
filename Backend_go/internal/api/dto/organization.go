package dto

import (
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/organization"
	"github.com/google/uuid"
)

// CreateOrganizationRequest represents the request body for creating a new organization
// @Description Request body for creating a new organization in the system
type CreateOrganizationRequest struct {
	Name        string                          `json:"name" binding:"required" example:"Acme Corporation"`
	Description string                          `json:"description" example:"A leading technology company"`
	Status      organization.OrganizationStatus `json:"status" example:"Active"`
}

// UpdateOrganizationRequest represents the request body for updating an existing organization
// @Description Request body for updating organization information
type UpdateOrganizationRequest struct {
	Name        *string                          `json:"name,omitempty" example:"Updated Acme Corp"`
	Description *string                          `json:"description,omitempty" example:"An innovative technology leader"`
	Status      *organization.OrganizationStatus `json:"status,omitempty" example:"Active"`
	OwnerID     *uuid.UUID                       `json:"owner_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// OrganizationResponse represents an organization in API responses
// @Description Detailed organization information returned in API responses
type OrganizationResponse struct {
	ID          uuid.UUID                       `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string                          `json:"name" example:"Acme Corporation"`
	Description string                          `json:"description" example:"A leading technology company"`
	Status      organization.OrganizationStatus `json:"status" example:"Active"`
	CreatedAt   time.Time                       `json:"created_at" example:"2024-03-15T09:00:00Z"`
	UpdatedAt   time.Time                       `json:"updated_at" example:"2024-03-15T10:30:00Z"`
	CreatorID   uuid.UUID                       `json:"creator_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OwnerID     uuid.UUID                       `json:"owner_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// OrganizationListResponse represents a paginated list of organizations
// @Description Paginated list of organizations with metadata
type OrganizationListResponse struct {
	Organizations []OrganizationResponse `json:"organizations"`
	TotalCount    int64                  `json:"total_count" example:"100"`
	Page          int                    `json:"page" example:"1"`
	PageSize      int                    `json:"page_size" example:"20"`
}

// OrganizationStatsResponse represents organization statistics
// @Description Statistical information about an organization
type OrganizationStatsResponse struct {
	Organization  OrganizationResponse `json:"organization"`
	MembersCount  int64                `json:"members_count" example:"50"`
	ProjectsCount int64                `json:"projects_count" example:"10"`
	TasksCount    int64                `json:"tasks_count" example:"200"`
}

// Convert domain Organization to OrganizationResponse
func OrganizationToResponse(org *organization.Organization) *OrganizationResponse {
	if org == nil {
		return nil
	}
	return &OrganizationResponse{
		ID:          org.ID,
		Name:        org.Name,
		Description: org.Description,
		Status:      org.Status,
		CreatedAt:   org.CreatedAt,
		UpdatedAt:   org.UpdatedAt,
		CreatorID:   org.CreatorID,
		OwnerID:     org.OwnerID,
	}
}

// Convert domain Organizations to OrganizationResponses
func OrganizationsToResponse(orgs []organization.Organization) []*OrganizationResponse {
	responses := make([]*OrganizationResponse, len(orgs))
	for i, org := range orgs {
		responses[i] = OrganizationToResponse(&org)
	}
	return responses
}
