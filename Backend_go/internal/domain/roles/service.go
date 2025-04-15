package roles

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

// Service interface for auth operations
type Service interface {
	// Role operations
	CreateRole(ctx context.Context, input CreateRoleInput) (*Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (*Role, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)
	ListRoles(ctx context.Context) ([]Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, input UpdateRoleInput) (*Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error

	// Permission operations
	CreatePermission(ctx context.Context, input CreatePermissionInput) (*Permission, error)
	GetPermission(ctx context.Context, id uuid.UUID) (*Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*Permission, error)
	ListPermissions(ctx context.Context) ([]Permission, error)
	UpdatePermission(ctx context.Context, id uuid.UUID, input UpdatePermissionInput) (*Permission, error)
	DeletePermission(ctx context.Context, id uuid.UUID) error

	// Role-Permission operations
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]Permission, error)

	// User-Role operations
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]Permission, error)
	UserHasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error)
	GetUserIDsByRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
}

type service struct {
	repo Repository
}

// NewService creates a new auth service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

// Input types
type CreateRoleInput struct {
	Name        string
	Description string
}

type UpdateRoleInput struct {
	Name        *string
	Description *string
}

type CreatePermissionInput struct {
	Name        string
	Description string
}

type UpdatePermissionInput struct {
	Name        *string
	Description *string
}

// Role operations implementation
func (s *service) CreateRole(ctx context.Context, input CreateRoleInput) (*Role, error) {
	if input.Name == "" {
		return nil, ErrInvalidInput
	}

	role := &Role{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *service) GetRole(ctx context.Context, id uuid.UUID) (*Role, error) {
	return s.repo.GetRole(ctx, id)
}

func (s *service) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.GetRoleByName(ctx, name)
}

func (s *service) ListRoles(ctx context.Context) ([]Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *service) UpdateRole(ctx context.Context, id uuid.UUID, input UpdateRoleInput) (*Role, error) {
	role, err := s.repo.GetRole(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		role.Name = *input.Name
	}
	if input.Description != nil {
		role.Description = *input.Description
	}

	if err := s.repo.UpdateRole(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *service) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteRole(ctx, id)
}

// Permission operations implementation
func (s *service) CreatePermission(ctx context.Context, input CreatePermissionInput) (*Permission, error) {
	if input.Name == "" {
		return nil, ErrInvalidInput
	}

	permission := &Permission{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := s.repo.CreatePermission(ctx, permission); err != nil {
		return nil, err
	}

	return permission, nil
}

func (s *service) GetPermission(ctx context.Context, id uuid.UUID) (*Permission, error) {
	return s.repo.GetPermission(ctx, id)
}

func (s *service) GetPermissionByName(ctx context.Context, name string) (*Permission, error) {
	if name == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.GetPermissionByName(ctx, name)
}

func (s *service) ListPermissions(ctx context.Context) ([]Permission, error) {
	return s.repo.ListPermissions(ctx)
}

func (s *service) UpdatePermission(ctx context.Context, id uuid.UUID, input UpdatePermissionInput) (*Permission, error) {
	permission, err := s.repo.GetPermission(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		permission.Name = *input.Name
	}
	if input.Description != nil {
		permission.Description = *input.Description
	}

	if err := s.repo.UpdatePermission(ctx, permission); err != nil {
		return nil, err
	}

	return permission, nil
}

func (s *service) DeletePermission(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePermission(ctx, id)
}

// Role-Permission operations implementation
func (s *service) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	// Verify role and permission exist
	if _, err := s.repo.GetRole(ctx, roleID); err != nil {
		return err
	}
	if _, err := s.repo.GetPermission(ctx, permissionID); err != nil {
		return err
	}

	return s.repo.AssignPermissionToRole(ctx, roleID, permissionID)
}

func (s *service) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return s.repo.RemovePermissionFromRole(ctx, roleID, permissionID)
}

func (s *service) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]Permission, error) {
	return s.repo.GetRolePermissions(ctx, roleID)
}

// User-Role operations implementation
func (s *service) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	// Verify role exists
	if _, err := s.repo.GetRole(ctx, roleID); err != nil {
		return err
	}

	return s.repo.AssignRoleToUser(ctx, userID, roleID)
}

func (s *service) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.repo.RemoveRoleFromUser(ctx, userID, roleID)
}

func (s *service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error) {
	return s.repo.GetUserRoles(ctx, userID)
}

func (s *service) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]Permission, error) {
	return s.repo.GetUserPermissions(ctx, userID)
}

// UserHasRole checks if a user has a specific role.
func (s *service) UserHasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error) {
	return s.repo.UserHasRole(ctx, userID, roleID)
}

// GetUserIDsByRole retrieves all user IDs associated with a specific role.
func (s *service) GetUserIDsByRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.GetUserIDsByRole(ctx, roleID)
}
