package project

import (
	"context"
	"errors"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/connection"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines the interface for project persistence operations
type Repository interface {
	Create(ctx context.Context, project *Project) error
	FindByID(ctx context.Context, id uuid.UUID) (*Project, error)
	FindAll(ctx context.Context, filter ProjectFilter) ([]Project, int64, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByName(ctx context.Context, name string, organizationID uuid.UUID) (*Project, error)
	AddMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, role string) error
	RemoveMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) error
}

type repository struct {
	db *connection.Database
}

func NewRepository(db *connection.Database) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, project *Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	var project Project
	result := r.db.WithContext(ctx).First(&project, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, result.Error
	}
	return &project, nil
}

func (r *repository) FindAll(ctx context.Context, filter ProjectFilter) ([]Project, int64, error) {
	var projects []Project
	var total int64
	query := r.db.WithContext(ctx).Model(&Project{})

	if filter.OrganizationID != nil {
		query = query.Where("organization_id = ?", filter.OrganizationID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Name != nil {
		query = query.Where("name LIKE ?", "%"+*filter.Name+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(filter.Page * filter.PageSize).
		Limit(filter.PageSize).
		Find(&projects).Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (r *repository) Update(ctx context.Context, project *Project) error {
	result := r.db.WithContext(ctx).Save(project)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Project{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (r *repository) FindByName(ctx context.Context, name string, organizationID uuid.UUID) (*Project, error) {
	var project Project
	result := r.db.WithContext(ctx).
		Where("name = ? AND organization_id = ?", name, organizationID).
		First(&project)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &project, nil
}

func (r *repository) AddMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, role string) error {
	member := struct {
		ProjectID uuid.UUID `gorm:"type:uuid;primary_key"`
		UserID    uuid.UUID `gorm:"type:uuid;primary_key"`
		Role      string
	}{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
	}
	return r.db.WithContext(ctx).Table("project_members").Create(&member).Error
}

func (r *repository) RemoveMember(ctx context.Context, projectID uuid.UUID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).Table("project_members").
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(nil)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrProjectNotFound
	}
	return nil
}
