package project

import (
	"context"
	"time"

	"github.com/google/uuid"
)



// Service interface
type Service interface {
	CreateProject(ctx context.Context, input CreateProjectInput) (*Project, error)
	GetProject(ctx context.Context, id uuid.UUID) (*Project, error)
	ListProjects(ctx context.Context, filter ProjectFilter) ([]Project, int64, error)
	UpdateProject(ctx context.Context, id uuid.UUID, input UpdateProjectInput) (*Project, error)
	DeleteProject(ctx context.Context, id uuid.UUID) error
	GetProjectDetails(ctx context.Context, id uuid.UUID) (*ProjectDetails, error)
	AddProjectMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, role string) error
	RemoveProjectMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) error
	UpdateProjectStatus(ctx context.Context, id uuid.UUID, status ProjectStatus) (*Project, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateProject(ctx context.Context, input CreateProjectInput) (*Project, error) {
	// Validate input
	if input.Name == "" {
		return nil, ErrInvalidInput
	}

	// Check if project name exists in organization
	existingProject, err := s.repo.FindByName(ctx, input.Name, input.OrganizationID)
	if err != nil {
		return nil, err
	}
	if existingProject != nil {
		return nil, ErrProjectNameExists
	}

	// Set default status if not provided
	if input.Status == "" {
		input.Status = ProjectStatusActive
	}

	project := &Project{
		ID:             uuid.New(),
		Name:           input.Name,
		Description:    input.Description,
		Status:         input.Status,
		CreatorID:      input.CreatorID,
		OrganizationID: input.OrganizationID,
		OwnerID:        input.OwnerID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = s.repo.Create(ctx, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *service) GetProject(ctx context.Context, id uuid.UUID) (*Project, error) {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}
	return project, nil
}

func (s *service) ListProjects(ctx context.Context, filter ProjectFilter) ([]Project, int64, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *service) UpdateProject(ctx context.Context, id uuid.UUID, input UpdateProjectInput) (*Project, error) {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	// Update fields if provided
	if input.Name != nil {
		// Check if new name exists in organization
		if *input.Name != project.Name {
			existingProject, err := s.repo.FindByName(ctx, *input.Name, project.OrganizationID)
			if err != nil {
				return nil, err
			}
			if existingProject != nil {
				return nil, ErrProjectNameExists
			}
		}
		project.Name = *input.Name
	}

	if input.Description != nil {
		project.Description = *input.Description
	}

	if input.Status != nil {
		if !input.Status.IsValid() {
			return nil, ErrInvalidInput
		}
		project.Status = *input.Status
	}

	if input.OwnerID != nil {
		project.OwnerID = *input.OwnerID
	}

	project.UpdatedAt = time.Now()
	err = s.repo.Update(ctx, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}

func (s *service) DeleteProject(ctx context.Context, id uuid.UUID) error {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if project == nil {
		return ErrProjectNotFound
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) GetProjectDetails(ctx context.Context, id uuid.UUID) (*ProjectDetails, error) {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	// TODO: Implement counting of members and tasks
	// This would require additional repository methods or joins
	details := &ProjectDetails{
		Project:      project,
		MembersCount: 0,                 // To be implemented with proper counting
		TasksCount:   0,                 // To be implemented with proper counting
		Members:      []ProjectMember{}, // To be implemented with proper member fetching
	}

	return details, nil
}

func (s *service) AddProjectMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, role string) error {
	project, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return ErrProjectNotFound
	}

	return s.repo.AddMember(ctx, projectID, userID, role)
}

func (s *service) RemoveProjectMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) error {
	project, err := s.repo.FindByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return ErrProjectNotFound
	}

	return s.repo.RemoveMember(ctx, projectID, userID)
}

func (s *service) UpdateProjectStatus(ctx context.Context, id uuid.UUID, status ProjectStatus) (*Project, error) {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}

	if !status.IsValid() {
		return nil, ErrInvalidInput
	}

	project.Status = status
	project.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, project)
	if err != nil {
		return nil, err
	}

	return project, nil
}
