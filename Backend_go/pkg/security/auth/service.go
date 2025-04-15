package auth

import (
	"context"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/roles"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/user"
	"github.com/google/uuid"
)

type Service interface {
	CreateRole(ctx context.Context, input roles.CreateRoleInput) (*roles.Role, error)
	GetRole(ctx context.Context, id uuid.UUID) (*roles.Role, error)
	ListRoles(ctx context.Context) ([]roles.Role, error)
	UpdateRole(ctx context.Context, id uuid.UUID, input roles.UpdateRoleInput) (*roles.Role, error)
	DeleteRole(ctx context.Context, id uuid.UUID) error
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]roles.Role, error)
	CreatePermission(ctx context.Context, input roles.CreatePermissionInput) (*roles.Permission, error)
	AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
}

type service struct {
	userRepo user.Repository
	rolesSvc roles.Service
}

func NewService(userRepo user.Repository, rolesSvc roles.Service) Service {
	return &service{
		userRepo: userRepo,
		rolesSvc: rolesSvc,
	}
}

func (s *service) CreateRole(ctx context.Context, input roles.CreateRoleInput) (*roles.Role, error) {
	return s.rolesSvc.CreateRole(ctx, input)
}

func (s *service) GetRole(ctx context.Context, id uuid.UUID) (*roles.Role, error) {
	return s.rolesSvc.GetRole(ctx, id)
}

func (s *service) ListRoles(ctx context.Context) ([]roles.Role, error) {
	return s.rolesSvc.ListRoles(ctx)
}

func (s *service) UpdateRole(ctx context.Context, id uuid.UUID, input roles.UpdateRoleInput) (*roles.Role, error) {
	return s.rolesSvc.UpdateRole(ctx, id, input)
}

func (s *service) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return s.rolesSvc.DeleteRole(ctx, id)
}

func (s *service) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.rolesSvc.AssignRoleToUser(ctx, userID, roleID)
}

func (s *service) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]roles.Role, error) {
	return s.rolesSvc.GetUserRoles(ctx, userID)
}

func (s *service) CreatePermission(ctx context.Context, input roles.CreatePermissionInput) (*roles.Permission, error) {
	return s.rolesSvc.CreatePermission(ctx, input)
}

func (s *service) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return s.rolesSvc.AssignPermissionToRole(ctx, roleID, permissionID)
}

func (s *service) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return s.rolesSvc.RemovePermissionFromRole(ctx, roleID, permissionID)
}
