package dto

import (
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/roles"
	"github.com/google/uuid"
)

// CreateRoleRequest represents the request body for creating a role
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required" example:"admin"`
	Description string `json:"description" example:"Administrator role with full access"`
}

// UpdateRoleRequest represents the request body for updating a role
type UpdateRoleRequest struct {
	Name        *string `json:"name" example:"admin"`
	Description *string `json:"description" example:"Administrator role with full access"`
}

// RoleResponse represents a role in API responses
type RoleResponse struct {
	ID          uuid.UUID            `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string               `json:"name" example:"admin"`
	Description string               `json:"description" example:"Administrator role with full access"`
	Permissions []PermissionResponse `json:"permissions"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// CreatePermissionRequest represents the request body for creating a permission
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required" example:"users:create"`
	Description string `json:"description" example:"Allows creating new users"`
}

// UpdatePermissionRequest represents the request body for updating a permission
type UpdatePermissionRequest struct {
	Name        *string `json:"name" example:"users:create"`
	Description *string `json:"description" example:"Allows creating new users"`
}

// PermissionResponse represents a permission in API responses
type PermissionResponse struct {
	ID          uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string    `json:"name" example:"users:create"`
	Description string    `json:"description" example:"Allows creating new users"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleToResponse converts a Role domain model to a RoleResponse DTO
func RoleToResponse(role *roles.Role) *RoleResponse {
	if role == nil {
		return nil
	}

	permissions := make([]PermissionResponse, len(role.Permissions))
	for i, perm := range role.Permissions {
		permissions[i] = *PermissionToResponse(&perm)
	}

	return &RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissions,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

// PermissionToResponse converts a Permission domain model to a PermissionResponse DTO
func PermissionToResponse(permission *roles.Permission) *PermissionResponse {
	if permission == nil {
		return nil
	}

	return &PermissionResponse{
		ID:          permission.ID,
		Name:        permission.Name,
		Description: permission.Description,
		CreatedAt:   permission.CreatedAt,
		UpdatedAt:   permission.UpdatedAt,
	}
}
