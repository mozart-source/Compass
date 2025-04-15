package task

import (
	"context"
	"errors"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/connection"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTaskNotFound = errors.New("task not found")
	ErrInvalidInput = errors.New("invalid input")
)

// TaskFilter defines filtering options for tasks
type TaskFilter struct {
	OrganizationID *uuid.UUID
	ProjectID      *uuid.UUID
	Status         *TaskStatus
	Priority       *TaskPriority
	AssigneeID     *uuid.UUID
	CreatorID      *uuid.UUID
	ReviewerID     *uuid.UUID
	StartDate      *time.Time
	EndDate        *time.Time
	DueDateStart   *time.Time
	DueDateEnd     *time.Time
	Page           int
	PageSize       int
}

// AnalyticsFilter defines filtering options for task analytics
type AnalyticsFilter struct {
	TaskID    *uuid.UUID
	UserID    *uuid.UUID
	Action    *string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int
}

// TaskRepository defines the interface for task persistence operations
type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id uuid.UUID) (*Task, error)
	FindAll(ctx context.Context, filter TaskFilter) ([]Task, int64, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Analytics methods
	RecordTaskActivity(ctx context.Context, analytics *TaskAnalytics) error
	GetTaskAnalytics(ctx context.Context, filter AnalyticsFilter) ([]TaskAnalytics, int64, error)
	GetTaskActivitySummary(ctx context.Context, taskID uuid.UUID, startTime, endTime time.Time) (map[string]int, error)
	GetUserTaskActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (map[string]int, error)
}

type taskRepository struct {
	db *connection.Database
}

func NewRepository(db *connection.Database) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) FindByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	var task Task
	result := r.db.WithContext(ctx).First(&task, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, result.Error
	}
	return &task, nil
}

func (r *taskRepository) FindAll(ctx context.Context, filter TaskFilter) ([]Task, int64, error) {
	var tasks []Task
	var total int64

	query := r.db.WithContext(ctx)

	// Apply filters
	if filter.OrganizationID != nil {
		query = query.Where("organization_id = ?", filter.OrganizationID)
	}
	if filter.ProjectID != nil {
		query = query.Where("project_id = ?", filter.ProjectID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Priority != nil {
		query = query.Where("priority = ?", filter.Priority)
	}
	if filter.AssigneeID != nil {
		query = query.Where("assignee_id = ?", filter.AssigneeID)
	}
	if filter.CreatorID != nil {
		query = query.Where("creator_id = ?", filter.CreatorID)
	}
	if filter.ReviewerID != nil {
		query = query.Where("reviewer_id = ?", filter.ReviewerID)
	}
	if filter.StartDate != nil && filter.EndDate != nil {
		query = query.Where("created_at BETWEEN ? AND ?", filter.StartDate, filter.EndDate)
	}
	if filter.DueDateStart != nil {
		query = query.Where("due_date >= ?", *filter.DueDateStart)
	}
	if filter.DueDateEnd != nil {
		query = query.Where("due_date < ?", *filter.DueDateEnd)
	}

	// Count total before pagination
	err := query.Model(&Task{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Set default PageSize if not set
	if filter.PageSize == 0 {
		filter.PageSize = 10000
	}

	// Apply pagination
	query = query.Offset(filter.Page * filter.PageSize).Limit(filter.PageSize)

	// Execute query
	if err := query.Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *taskRepository) Update(ctx context.Context, task *Task) error {
	result := r.db.WithContext(ctx).Save(task)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *taskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Task{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

// Analytics implementation
func (r *taskRepository) RecordTaskActivity(ctx context.Context, analytics *TaskAnalytics) error {
	return r.db.WithContext(ctx).Create(analytics).Error
}

func (r *taskRepository) GetTaskAnalytics(ctx context.Context, filter AnalyticsFilter) ([]TaskAnalytics, int64, error) {
	var analytics []TaskAnalytics
	var total int64
	query := r.db.WithContext(ctx).Model(&TaskAnalytics{})

	if filter.TaskID != nil {
		query = query.Where("task_id = ?", *filter.TaskID)
	}
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

func (r *taskRepository) GetTaskActivitySummary(ctx context.Context, taskID uuid.UUID, startTime, endTime time.Time) (map[string]int, error) {
	var results []struct {
		Action string
		Count  int
	}

	err := r.db.WithContext(ctx).Model(&TaskAnalytics{}).
		Select("action, count(*) as count").
		Where("task_id = ? AND timestamp BETWEEN ? AND ?", taskID, startTime, endTime).
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

func (r *taskRepository) GetUserTaskActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (map[string]int, error) {
	var results []struct {
		Action string
		Count  int
	}

	err := r.db.WithContext(ctx).Model(&TaskAnalytics{}).
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
