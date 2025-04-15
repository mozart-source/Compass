package organization

import (
	"context"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/connection"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository defines the interface for organization data access
type Repository interface {
	Create(ctx context.Context, org *Organization) error
	FindByID(ctx context.Context, id uuid.UUID) (*Organization, error)
	FindAll(ctx context.Context, filter OrganizationFilter) ([]Organization, int64, error)
	Update(ctx context.Context, org *Organization) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByName(ctx context.Context, name string) (*Organization, error)
}

// OrganizationFilter represents the filter options for listing organizations
type OrganizationFilter struct {
	Page     int
	PageSize int
}

type repository struct {
	db *gorm.DB
}

// NewRepository creates a new organization repository
func NewRepository(db *connection.Database) Repository {
	return &repository{
		db: db.DB,
	}
}

// Create creates a new organization in the database
func (r *repository) Create(ctx context.Context, org *Organization) error {
	result := r.db.WithContext(ctx).Create(org)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// FindByID retrieves an organization by its ID
func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*Organization, error) {
	var org Organization
	result := r.db.WithContext(ctx).First(&org, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrOrganizationNotFound
		}
		return nil, result.Error
	}
	return &org, nil
}

// FindAll retrieves all organizations with pagination
func (r *repository) FindAll(ctx context.Context, filter OrganizationFilter) ([]Organization, int64, error) {
	var organizations []Organization
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&Organization{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated records
	result := r.db.WithContext(ctx).
		Offset(filter.Page * filter.PageSize).
		Limit(filter.PageSize).
		Find(&organizations)

	if result.Error != nil {
		return nil, 0, result.Error
	}

	return organizations, total, nil
}

// Update updates an existing organization
func (r *repository) Update(ctx context.Context, org *Organization) error {
	result := r.db.WithContext(ctx).Save(org)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOrganizationNotFound
	}
	return nil
}

// Delete deletes an organization by its ID
func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Organization{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOrganizationNotFound
	}
	return nil
}

// FindByName retrieves an organization by its name
func (r *repository) FindByName(ctx context.Context, name string) (*Organization, error) {
	var org Organization
	result := r.db.WithContext(ctx).First(&org, "name = ?", name)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &org, nil
}
