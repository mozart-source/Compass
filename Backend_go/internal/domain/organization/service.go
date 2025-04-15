package organization

import (
	"context"

	"github.com/google/uuid"
)

// Input types
type CreateOrganizationInput struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Status      OrganizationStatus `json:"status"`
	CreatorID   uuid.UUID          `json:"creator_id"`
	OwnerID     uuid.UUID          `json:"owner_id"`
}

type UpdateOrganizationInput struct {
	Name        *string             `json:"name,omitempty"`
	Description *string             `json:"description,omitempty"`
	Status      *OrganizationStatus `json:"status,omitempty"`
	OwnerID     *uuid.UUID          `json:"owner_id,omitempty"`
}

// Service defines the interface for organization business logic
type Service interface {
	CreateOrganization(ctx context.Context, input CreateOrganizationInput) (*Organization, error)
	GetOrganization(ctx context.Context, id uuid.UUID) (*Organization, error)
	ListOrganizations(ctx context.Context, filter OrganizationFilter) ([]Organization, int64, error)
	UpdateOrganization(ctx context.Context, id uuid.UUID, input UpdateOrganizationInput) (*Organization, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
	GetOrganizationByName(ctx context.Context, name string) (*Organization, error)
}

type service struct {
	repo Repository
}

// NewService creates a new organization service instance
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// CreateOrganization creates a new organization
func (s *service) CreateOrganization(ctx context.Context, input CreateOrganizationInput) (*Organization, error) {
	// Validate input
	if input.Name == "" {
		return nil, ErrInvalidInput
	}
	if input.CreatorID == uuid.Nil {
		return nil, ErrInvalidCreator
	}
	if input.OwnerID == uuid.Nil {
		return nil, ErrInvalidOwner
	}

	// Check if organization name exists
	existingOrg, err := s.repo.FindByName(ctx, input.Name)
	if err != nil {
		return nil, err
	}
	if existingOrg != nil {
		return nil, ErrDuplicateName
	}

	// Set default status if not provided
	if input.Status == "" {
		input.Status = OrganizationStatusActive
	}

	// Create organization
	org := &Organization{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		Status:      input.Status,
		CreatorID:   input.CreatorID,
		OwnerID:     input.OwnerID,
	}

	if err := s.repo.Create(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

// GetOrganization retrieves an organization by ID
func (s *service) GetOrganization(ctx context.Context, id uuid.UUID) (*Organization, error) {
	org, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return org, nil
}

// ListOrganizations retrieves all organizations with pagination
func (s *service) ListOrganizations(ctx context.Context, filter OrganizationFilter) ([]Organization, int64, error) {
	// Validate pagination parameters
	if filter.Page < 0 {
		filter.Page = 0
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	return s.repo.FindAll(ctx, filter)
}

// UpdateOrganization updates an existing organization
func (s *service) UpdateOrganization(ctx context.Context, id uuid.UUID, input UpdateOrganizationInput) (*Organization, error) {
	// Get existing organization
	org, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check name uniqueness if name is being updated
	if input.Name != nil && *input.Name != org.Name {
		existingOrg, err := s.repo.FindByName(ctx, *input.Name)
		if err != nil {
			return nil, err
		}
		if existingOrg != nil {
			return nil, ErrDuplicateName
		}
		org.Name = *input.Name
	}

	// Update fields if provided
	if input.Description != nil {
		org.Description = *input.Description
	}

	if input.Status != nil {
		if !input.Status.IsValid() {
			return nil, ErrInvalidStatus
		}
		org.Status = *input.Status
	}

	if input.OwnerID != nil {
		if *input.OwnerID == uuid.Nil {
			return nil, ErrInvalidOwner
		}
		org.OwnerID = *input.OwnerID
	}

	// Save changes
	if err := s.repo.Update(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

// DeleteOrganization deletes an organization
func (s *service) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	// Check if organization exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	return s.repo.Delete(ctx, id)
}

// GetOrganizationByName retrieves an organization by name
func (s *service) GetOrganizationByName(ctx context.Context, name string) (*Organization, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}

	org, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, ErrOrganizationNotFound
	}

	return org, nil
}
