package todos

import (
	"context"
	"errors"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/persistence/postgres/connection"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrTodoNotFound = errors.New("todo not found")
)

type TodoRepository interface {
	Create(ctx context.Context, todo *Todo) error
	FindByID(ctx context.Context, id uuid.UUID) (*Todo, error)
	FindAll(ctx context.Context, filter TodoFilter) ([]Todo, int64, error)
	Update(ctx context.Context, todo *Todo) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]Todo, error)
	FindByListID(ctx context.Context, listID uuid.UUID) ([]Todo, error)
	FindByUserIDAndListID(ctx context.Context, userID uuid.UUID, listID uuid.UUID) ([]Todo, error)
	FindCompletedByUserID(ctx context.Context, userID uuid.UUID) ([]Todo, error)
	FindUncompletedByUserID(ctx context.Context, userID uuid.UUID) ([]Todo, error)
	CreateTodoList(ctx context.Context, list *TodoList) error
	GetOrCreateDefaultList(ctx context.Context, userID uuid.UUID) (*TodoList, error)
	FindDefaultListByUserID(ctx context.Context, userID uuid.UUID) (*TodoList, error)
	UpdateTodoList(ctx context.Context, list *TodoList) error
	DeleteTodoList(ctx context.Context, id uuid.UUID) error
	FindTodoListByID(ctx context.Context, id uuid.UUID) (*TodoList, error)
	FindAllTodoLists(ctx context.Context, userID uuid.UUID) ([]TodoList, error)
}

type todoRepository struct {
	db *connection.Database
}

func NewTodoRepository(db *connection.Database) TodoRepository {
	return &todoRepository{db: db}
}

func (r *todoRepository) Create(ctx context.Context, todo *Todo) error {
	return r.db.WithContext(ctx).Create(todo).Error
}

func (r *todoRepository) FindByID(ctx context.Context, id uuid.UUID) (*Todo, error) {
	var todo Todo
	result := r.db.WithContext(ctx).First(&todo, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTodoNotFound
		}
		return nil, result.Error
	}
	return &todo, nil
}

func (r *todoRepository) FindAll(ctx context.Context, filter TodoFilter) ([]Todo, int64, error) {
	var todos []Todo
	var total int64

	query := r.db.WithContext(ctx)

	// Apply filters
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Priority != nil {
		query = query.Where("priority = ?", filter.Priority)
	}
	if filter.IsCompleted != nil {
		query = query.Where("is_completed = ?", *filter.IsCompleted)
	}
	if filter.DueDateStart != nil {
		query = query.Where("due_date >= ?", *filter.DueDateStart)
	}
	if filter.DueDateEnd != nil {
		query = query.Where("due_date < ?", *filter.DueDateEnd)
	}
	if filter.DueDate != nil {
		query = query.Where("due_date = ?", filter.DueDate)
	}
	if filter.ReminderTime != nil {
		query = query.Where("reminder_time = ?", filter.ReminderTime)
	}
	if filter.IsRecurring != nil {
		query = query.Where("is_recurring = ?", filter.IsRecurring)
	}
	if filter.Tags != nil {
		query = query.Where("tags = ?", filter.Tags)
	}
	if filter.Checklist != nil {
		query = query.Where("checklist = ?", filter.Checklist)
	}
	if filter.LinkedTaskID != nil {
		query = query.Where("linked_task_id = ?", filter.LinkedTaskID)
	}
	if filter.LinkedCalendarEventID != nil {
		query = query.Where("linked_calendar_event_id = ?", filter.LinkedCalendarEventID)
	}

	// Count total before pagination
	err := query.Model(&Todo{}).Count(&total).Error
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
	if err := query.Find(&todos).Error; err != nil {
		return nil, 0, err
	}

	return todos, total, nil
}

func (r *todoRepository) Update(ctx context.Context, todo *Todo) error {
	result := r.db.WithContext(ctx).Save(todo)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTodoNotFound
	}
	return nil
}

func (r *todoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&Todo{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTodoNotFound
	}
	return nil
}

func (r *todoRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]Todo, error) {
	var todos []Todo
	result := r.db.WithContext(ctx).
		Model(&Todo{}).
		Where("user_id = ?", userID).
		Find(&todos)

	if result.Error != nil {
		return nil, result.Error
	}

	// Initialize empty maps for any nil JSONB fields
	for i := range todos {
		if todos[i].RecurrencePattern == nil {
			todos[i].RecurrencePattern = make(map[string]interface{})
		}
		if todos[i].Tags == nil {
			todos[i].Tags = make(map[string]interface{})
		}
		if todos[i].Checklist == nil {
			todos[i].Checklist = make(map[string]interface{})
		}
		if todos[i].AISuggestions == nil {
			todos[i].AISuggestions = make(map[string]interface{})
		}
	}

	return todos, nil
}

func (r *todoRepository) FindByListID(ctx context.Context, listID uuid.UUID) ([]Todo, error) {
	var todos []Todo
	result := r.db.WithContext(ctx).Where("list_id = ?", listID).Find(&todos)
	if result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

func (r *todoRepository) FindByUserIDAndListID(ctx context.Context, userID uuid.UUID, listID uuid.UUID) ([]Todo, error) {
	var todos []Todo
	result := r.db.WithContext(ctx).Where("user_id = ? AND list_id = ?", userID, listID).Find(&todos)
	if result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

func (r *todoRepository) FindCompletedByUserID(ctx context.Context, userID uuid.UUID) ([]Todo, error) {
	var todos []Todo
	result := r.db.WithContext(ctx).Where("user_id = ? AND is_completed = true", userID).Find(&todos)
	if result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

func (r *todoRepository) FindUncompletedByUserID(ctx context.Context, userID uuid.UUID) ([]Todo, error) {
	var todos []Todo
	result := r.db.WithContext(ctx).Where("user_id = ? AND is_completed = false", userID).Find(&todos)
	if result.Error != nil {
		return nil, result.Error
	}
	return todos, nil
}

func (r *todoRepository) CreateTodoList(ctx context.Context, list *TodoList) error {
	return r.db.WithContext(ctx).Create(list).Error
}

func (r *todoRepository) GetOrCreateDefaultList(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	// First try to find existing default list
	list, err := r.FindDefaultListByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// If found, return it
	if list != nil {
		return list, nil
	}

	// Create new default list
	defaultList := &TodoList{
		ID:          uuid.New(),
		UserID:      userID,
		Name:        "Default List",
		Description: "Default todo list",
		IsDefault:   true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := r.CreateTodoList(ctx, defaultList); err != nil {
		return nil, err
	}

	return defaultList, nil
}

func (r *todoRepository) FindDefaultListByUserID(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	var list TodoList
	result := r.db.WithContext(ctx).Where("user_id = ? AND is_default = true", userID).First(&list)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, result.Error
		}
		return nil, result.Error
	}
	return &list, nil
}

func (r *todoRepository) UpdateTodoList(ctx context.Context, list *TodoList) error {
	result := r.db.WithContext(ctx).Save(list)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTodoNotFound
	}
	return nil
}

func (r *todoRepository) DeleteTodoList(ctx context.Context, id uuid.UUID) error {
	// Start a transaction
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Delete all todos associated with this list first
	if err := tx.Where("list_id = ?", id).Delete(&Todo{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Then delete the list itself
	if err := tx.Delete(&TodoList{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *todoRepository) FindTodoListByID(ctx context.Context, id uuid.UUID) (*TodoList, error) {
	var list TodoList
	result := r.db.WithContext(ctx).Preload("Todos").First(&list, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrTodoNotFound
		}
		return nil, result.Error
	}
	return &list, nil
}

func (r *todoRepository) FindAllTodoLists(ctx context.Context, userID uuid.UUID) ([]TodoList, error) {
	var lists []TodoList
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Preload("Todos").Find(&lists)
	if result.Error != nil {
		return nil, result.Error
	}
	return lists, nil
}
