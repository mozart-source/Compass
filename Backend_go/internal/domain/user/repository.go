package user

import (
	"context"
	"errors"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/connection"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidInput = errors.New("invalid input")
)

// UserFilter defines the filtering options for users
type UserFilter struct {
	IsActive    *bool
	Email       *string
	Username    *string
	FirstName   *string
	LastName    *string
	PhoneNumber *string
	Timezone    *string
	Locale      *string
	Page        int
	PageSize    int
}

// AnalyticsFilter defines filtering options for user analytics
type AnalyticsFilter struct {
	UserID    *uuid.UUID
	Action    *string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int
}

type Repository interface {
	Create(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id uuid.UUID) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByProviderID(ctx context.Context, providerID, provider string) (*User, error)
	FindAll(ctx context.Context, filter UserFilter) ([]User, int64, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Analytics methods
	RecordUserActivity(ctx context.Context, analytics *UserAnalytics) error
	RecordSessionActivity(ctx context.Context, analytics *SessionAnalytics) error
	GetUserAnalytics(ctx context.Context, filter AnalyticsFilter) ([]UserAnalytics, int64, error)
	GetSessionAnalytics(ctx context.Context, filter AnalyticsFilter) ([]SessionAnalytics, int64, error)
	GetUserActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (map[string]int, error)
	CountLogins(ctx context.Context, userID uuid.UUID) (int, error)
	CountActions(ctx context.Context, userID uuid.UUID) (int, error)
}

type repository struct {
	db *connection.Database
}

func NewRepository(db *connection.Database) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	var user User
	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *repository) FindAll(ctx context.Context, filter UserFilter) ([]User, int64, error) {
	var users []User
	var total int64
	query := r.db.WithContext(ctx).Model(&User{})

	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}
	if filter.Email != nil {
		query = query.Where("email LIKE ?", "%"+*filter.Email+"%")
	}
	if filter.Username != nil {
		query = query.Where("username LIKE ?", "%"+*filter.Username+"%")
	}
	if filter.FirstName != nil {
		query = query.Where("first_name LIKE ?", "%"+*filter.FirstName+"%")
	}
	if filter.LastName != nil {
		query = query.Where("last_name LIKE ?", "%"+*filter.LastName+"%")
	}
	if filter.PhoneNumber != nil {
		query = query.Where("phone_number LIKE ?", "%"+*filter.PhoneNumber+"%")
	}
	if filter.Timezone != nil {
		query = query.Where("timezone = ?", *filter.Timezone)
	}
	if filter.Locale != nil {
		query = query.Where("locale = ?", *filter.Locale)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Offset(filter.Page * filter.PageSize).
		Limit(filter.PageSize).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *repository) Update(ctx context.Context, user *User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&User{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	result := r.db.WithContext(ctx).Where("username = ?", username).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *repository) FindByProviderID(ctx context.Context, providerID, provider string) (*User, error) {
	var user User
	result := r.db.WithContext(ctx).Where("provider_id = ? AND provider = ?", providerID, provider).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// Analytics implementation
func (r *repository) RecordUserActivity(ctx context.Context, analytics *UserAnalytics) error {
	// Ensure metadata is valid JSON
	if analytics.Metadata == "" {
		analytics.Metadata = "{}"
	}

	return r.db.WithContext(ctx).Create(analytics).Error
}

func (r *repository) RecordSessionActivity(ctx context.Context, analytics *SessionAnalytics) error {
	return r.db.WithContext(ctx).Create(analytics).Error
}

func (r *repository) GetUserAnalytics(ctx context.Context, filter AnalyticsFilter) ([]UserAnalytics, int64, error) {
	var analytics []UserAnalytics
	var total int64
	query := r.db.WithContext(ctx).Model(&UserAnalytics{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Action != nil {
		query = query.Where("action = ?", *filter.Action)
	}
	if filter.StartTime != nil && filter.EndTime != nil {
		query = query.Where("timestamp BETWEEN ? AND ?", *filter.StartTime, *filter.EndTime)
	} else if filter.StartTime != nil {
		query = query.Where("timestamp >= ?", *filter.StartTime)
	} else if filter.EndTime != nil {
		query = query.Where("timestamp <= ?", *filter.EndTime)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("timestamp DESC").
		Offset(filter.Page * filter.PageSize).
		Limit(filter.PageSize).
		Find(&analytics).Error
	if err != nil {
		return nil, 0, err
	}

	return analytics, total, nil
}

func (r *repository) GetSessionAnalytics(ctx context.Context, filter AnalyticsFilter) ([]SessionAnalytics, int64, error) {
	var analytics []SessionAnalytics
	var total int64
	query := r.db.WithContext(ctx).Model(&SessionAnalytics{})

	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Action != nil {
		query = query.Where("action = ?", *filter.Action)
	}
	if filter.StartTime != nil && filter.EndTime != nil {
		query = query.Where("timestamp BETWEEN ? AND ?", *filter.StartTime, *filter.EndTime)
	} else if filter.StartTime != nil {
		query = query.Where("timestamp >= ?", *filter.StartTime)
	} else if filter.EndTime != nil {
		query = query.Where("timestamp <= ?", *filter.EndTime)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Order("timestamp DESC").
		Offset(filter.Page * filter.PageSize).
		Limit(filter.PageSize).
		Find(&analytics).Error
	if err != nil {
		return nil, 0, err
	}

	return analytics, total, nil
}

func (r *repository) GetUserActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (map[string]int, error) {
	var results []struct {
		Action string
		Count  int
	}

	err := r.db.WithContext(ctx).Model(&UserAnalytics{}).
		Select("action, count(*) as count").
		Where("user_id = ? AND timestamp BETWEEN ? AND ?", userID, startTime, endTime).
		Group("action").
		Find(&results).Error

	if err != nil {
		return nil, err
	}

	summary := make(map[string]int)
	for _, result := range results {
		summary[result.Action] = result.Count
	}

	return summary, nil
}

func (r *repository) CountLogins(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&UserAnalytics{}).
		Where("user_id = ? AND action = ?", userID, "login_success").
		Count(&count).Error
	return int(count), err
}

func (r *repository) CountActions(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&UserAnalytics{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return int(count), err
}
