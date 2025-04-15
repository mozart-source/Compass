package roles

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrRoleNotFound        = errors.New("role not found")
	ErrPermissionNotFound  = errors.New("permission not found")
	ErrDuplicateRole       = errors.New("role already exists")
	ErrDuplicatePermission = errors.New("permission already exists")
)

// Repository interface for auth operations
type Repository interface {
	// Role operations
	CreateRole(ctx context.Context, role *Role) error
	GetRole(ctx context.Context, id uuid.UUID) (*Role, error)
	GetRoleByName(ctx context.Context, name string) (*Role, error)
	ListRoles(ctx context.Context) ([]Role, error)
	UpdateRole(ctx context.Context, role *Role) error
	DeleteRole(ctx context.Context, id uuid.UUID) error

	// Permission operations
	CreatePermission(ctx context.Context, permission *Permission) error
	GetPermission(ctx context.Context, id uuid.UUID) (*Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*Permission, error)
	ListPermissions(ctx context.Context) ([]Permission, error)
	UpdatePermission(ctx context.Context, permission *Permission) error
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

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new auth repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// Role operations implementation
func (r *repository) CreateRole(ctx context.Context, role *Role) error {
	result := r.db.WithContext(ctx).Create(role)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *repository) GetRole(ctx context.Context, id uuid.UUID) (*Role, error) {
	var role Role
	result := r.db.WithContext(ctx).Preload("Permissions").First(&role, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, result.Error
	}
	return &role, nil
}

func (r *repository) GetRoleByName(ctx context.Context, name string) (*Role, error) {
	var role Role
	result := r.db.WithContext(ctx).Preload("Permissions").First(&role, "name = ?", name)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, result.Error
	}
	return &role, nil
}

func (r *repository) ListRoles(ctx context.Context) ([]Role, error) {
	var roles []Role
	result := r.db.WithContext(ctx).Preload("Permissions").Find(&roles)
	if result.Error != nil {
		return nil, result.Error
	}
	return roles, nil
}

func (r *repository) UpdateRole(ctx context.Context, role *Role) error {
	result := r.db.WithContext(ctx).Save(role)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *repository) DeleteRole(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Role{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRoleNotFound
	}
	return nil
}

// Permission operations implementation
func (r *repository) CreatePermission(ctx context.Context, permission *Permission) error {
	result := r.db.WithContext(ctx).Create(permission)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *repository) GetPermission(ctx context.Context, id uuid.UUID) (*Permission, error) {
	var permission Permission
	result := r.db.WithContext(ctx).First(&permission, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, result.Error
	}
	return &permission, nil
}

func (r *repository) GetPermissionByName(ctx context.Context, name string) (*Permission, error) {
	var permission Permission
	result := r.db.WithContext(ctx).First(&permission, "name = ?", name)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrPermissionNotFound
		}
		return nil, result.Error
	}
	return &permission, nil
}

func (r *repository) ListPermissions(ctx context.Context) ([]Permission, error) {
	var permissions []Permission
	result := r.db.WithContext(ctx).Find(&permissions)
	if result.Error != nil {
		return nil, result.Error
	}
	return permissions, nil
}

func (r *repository) UpdatePermission(ctx context.Context, permission *Permission) error {
	result := r.db.WithContext(ctx).Save(permission)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *repository) DeletePermission(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Permission{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPermissionNotFound
	}
	return nil
}

// Role-Permission operations implementation
func (r *repository) AssignPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).Create(&RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}).Error
}

func (r *repository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&RolePermission{}).Error
}

func (r *repository) GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]Permission, error) {
	var permissions []Permission
	err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// User-Role operations implementation
func (r *repository) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).Create(&UserRole{
		UserID: userID,
		RoleID: roleID,
	}).Error
}

func (r *repository) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&UserRole{}).Error
}

func (r *repository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]Role, error) {
	var roles []Role
	err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Preload("Permissions").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *repository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]Permission, error) {
	var permissions []Permission
	err := r.db.WithContext(ctx).
		Distinct().
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *repository) UserHasRole(ctx context.Context, userID, roleID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserRole{}).Where("user_id = ? AND role_id = ?", userID, roleID).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) GetUserIDsByRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID
	err := r.db.WithContext(ctx).Model(&UserRole{}).Where("role_id = ?", roleID).Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}
	return userIDs, nil
}
