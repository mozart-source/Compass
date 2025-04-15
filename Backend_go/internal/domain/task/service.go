package task

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/events"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/cache"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrInvalidTransition = errors.New("invalid status transition")
	ErrDependencyFailed  = errors.New("dependencies not completed")
)

// Analytics types
type RecordTaskActivityInput struct {
	TaskID    uuid.UUID              `json:"task_id"`
	UserID    uuid.UUID              `json:"user_id"`
	Action    string                 `json:"action"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
}

type TaskActivitySummary struct {
	TaskID       uuid.UUID      `json:"task_id"`
	ActionCounts map[string]int `json:"action_counts"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	TotalActions int            `json:"total_actions"`
}

type UserTaskActivitySummary struct {
	UserID       uuid.UUID      `json:"user_id"`
	ActionCounts map[string]int `json:"action_counts"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	TotalActions int            `json:"total_actions"`
}

type Service interface {
	CreateTask(ctx context.Context, input CreateTaskInput) (*Task, error)
	GetTask(ctx context.Context, id uuid.UUID) (*Task, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]Task, int64, error)
	UpdateTask(ctx context.Context, id uuid.UUID, input UpdateTaskInput) (*Task, error)
	UpdateTaskStatus(ctx context.Context, id uuid.UUID, status TaskStatus) (*Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID) error
	GetTaskMetrics(ctx context.Context, id uuid.UUID) (*TaskMetrics, error)
	GetProjectTasks(ctx context.Context, projectID uuid.UUID, filter TaskFilter) ([]Task, int64, error)
	AssignTask(ctx context.Context, id uuid.UUID, assigneeID uuid.UUID) (*Task, error)

	// Analytics methods
	RecordTaskActivity(ctx context.Context, input RecordTaskActivityInput) error
	GetTaskAnalytics(ctx context.Context, taskID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]TaskAnalytics, int64, error)
	GetUserTaskAnalytics(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]TaskAnalytics, int64, error)
	GetTaskActivitySummary(ctx context.Context, taskID uuid.UUID, startTime, endTime time.Time) (*TaskActivitySummary, error)
	GetUserTaskActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (*UserTaskActivitySummary, error)
	GetDashboardMetrics(userID uuid.UUID) (TasksDashboardMetrics, error)
	GetTodayTasks(ctx context.Context, userID uuid.UUID) ([]Task, error)
}

type TaskMetrics struct {
	HealthScore     float64                `json:"health_score"`
	ComplexityScore float64                `json:"complexity_score"`
	ProgressMetrics map[string]interface{} `json:"progress_metrics"`
	Blockers        []string               `json:"blockers"`
	RiskFactors     map[string]interface{} `json:"risk_factors"`
}

type CreateTaskInput struct {
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	Status         TaskStatus   `json:"status"`
	Priority       TaskPriority `json:"priority"`
	CreatorID      uuid.UUID    `json:"creator_id"`
	AssigneeID     *uuid.UUID   `json:"assignee_id,omitempty"`
	ReviewerID     *uuid.UUID   `json:"reviewer_id,omitempty"`
	CategoryID     *uuid.UUID   `json:"category_id,omitempty"`
	ParentTaskID   *uuid.UUID   `json:"parent_task_id,omitempty"`
	ProjectID      uuid.UUID    `json:"project_id"`
	OrganizationID uuid.UUID    `json:"organization_id"`
	EstimatedHours float64      `json:"estimated_hours,omitempty"`
	StartDate      time.Time    `json:"start_date"`
	Duration       *float64     `json:"duration,omitempty"`
	DueDate        *time.Time   `json:"due_date,omitempty"`
	Dependencies   []uuid.UUID  `json:"dependencies,omitempty"`
}

type UpdateTaskInput struct {
	Title          *string       `json:"title,omitempty"`
	Description    *string       `json:"description,omitempty"`
	Status         *TaskStatus   `json:"status,omitempty"`
	Priority       *TaskPriority `json:"priority,omitempty"`
	AssigneeID     *uuid.UUID    `json:"assignee_id,omitempty"`
	ReviewerID     *uuid.UUID    `json:"reviewer_id,omitempty"`
	CategoryID     *uuid.UUID    `json:"category_id,omitempty"`
	EstimatedHours *float64      `json:"estimated_hours,omitempty"`
	StartDate      *time.Time    `json:"start_date,omitempty"`
	Duration       *float64      `json:"duration,omitempty"`
	DueDate        *time.Time    `json:"due_date,omitempty"`
	Dependencies   []uuid.UUID   `json:"dependencies,omitempty"`
}

// Define TasksDashboardMetrics struct for dashboard metrics aggregation
// TasksDashboardMetrics represents summary metrics for the dashboard
// Used by GetDashboardMetrics
type TasksDashboardMetrics struct {
	Total     int
	Completed int
	Overdue   int
}

// Repository interface

type service struct {
	repo   TaskRepository
	redis  *cache.RedisClient // Injected for event publishing
	logger *zap.Logger
}

func NewService(repo TaskRepository, redis *cache.RedisClient, logger *zap.Logger) Service {
	return &service{repo: repo, redis: redis, logger: logger}
}

func (s *service) CreateTask(ctx context.Context, input CreateTaskInput) (*Task, error) {
	// Validate input
	if input.Title == "" {
		return nil, ErrInvalidInput
	}

	// Set default values
	if input.Status == "" {
		input.Status = TaskStatusUpcoming
	}
	if input.Priority == "" {
		input.Priority = TaskPriorityMedium
	}

	task := &Task{
		ID:             uuid.New(),
		Title:          input.Title,
		Description:    input.Description,
		Status:         input.Status,
		Priority:       input.Priority,
		CreatorID:      input.CreatorID,
		AssigneeID:     input.AssigneeID,
		ReviewerID:     input.ReviewerID,
		CategoryID:     input.CategoryID,
		ParentTaskID:   input.ParentTaskID,
		ProjectID:      input.ProjectID,
		OrganizationID: input.OrganizationID,
		EstimatedHours: input.EstimatedHours,
		StartDate:      input.StartDate,
		Duration:       input.Duration,
		DueDate:        input.DueDate,
		Dependencies:   input.Dependencies,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := s.repo.Create(ctx, task)
	if err != nil {
		return nil, err
	}

	s.recordTaskActivity(ctx, task, task.CreatorID, "task_created", map[string]interface{}{
		"title":  task.Title,
		"status": task.Status,
	})

	// Record task creation activity with meaningful metadata
	if callerID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		metadata := marshalTaskMetadata(map[string]interface{}{
			"created_by":      callerID.String(),
			"task_id":         task.ID.String(),
			"title":           task.Title,
			"status":          string(task.Status),
			"priority":        string(task.Priority),
			"project_id":      task.ProjectID.String(),
			"organization_id": task.OrganizationID.String(),
		})
		analytics := &TaskAnalytics{
			ID:        uuid.New(),
			TaskID:    task.ID,
			UserID:    callerID,
			Action:    "task_created",
			Timestamp: time.Now(),
			Metadata:  metadata,
		}
		_ = s.repo.RecordTaskActivity(ctx, analytics)
	} else {
		analytics := &TaskAnalytics{
			ID:        uuid.New(),
			TaskID:    task.ID,
			UserID:    task.CreatorID,
			Action:    "task_created",
			Timestamp: time.Now(),
			Metadata: marshalTaskMetadata(map[string]interface{}{
				"task_id": task.ID.String(),
				"title":   task.Title,
			}),
		}
		_ = s.repo.RecordTaskActivity(ctx, analytics)
	}

	// Publish dashboard event
	event := &events.DashboardEvent{
		EventType: events.DashboardEventCacheInvalidate,
		UserID:    input.CreatorID,
		Timestamp: time.Now().UTC(),
		Details: map[string]interface{}{
			"action":  "task_created",
			"task_id": task.ID,
		},
	}
	if err := s.redis.PublishDashboardEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish dashboard event", zap.Error(err))
	}

	return task, nil
}

func (s *service) GetTask(ctx context.Context, id uuid.UUID) (*Task, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}
	return task, nil
}

func (s *service) ListTasks(ctx context.Context, filter TaskFilter) ([]Task, int64, error) {
	return s.repo.FindAll(ctx, filter)
}

// Helper to marshal metadata
func marshalTaskMetadata(data map[string]interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func (s *service) UpdateTask(ctx context.Context, id uuid.UUID, input UpdateTaskInput) (*Task, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	// Store old values for change tracking
	oldStatus := task.Status
	oldAssignee := task.AssigneeID
	oldDependencies := task.Dependencies

	changed := false
	var analyticsEvents []*TaskAnalytics
	var callerID uuid.UUID
	if v, ok := ctx.Value("user_id").(uuid.UUID); ok {
		callerID = v
	} else {
		callerID = task.CreatorID
	}

	// Update fields if provided
	if input.Title != nil && *input.Title != task.Title {
		task.Title = *input.Title
		changed = true
	}
	if input.Description != nil && *input.Description != task.Description {
		task.Description = *input.Description
		changed = true
	}
	if input.Status != nil && *input.Status != oldStatus {
		task.Status = *input.Status
		changed = true
		metadata := marshalTaskMetadata(map[string]interface{}{
			"old_status": string(oldStatus),
			"new_status": string(*input.Status),
			"updated_by": callerID.String(),
			"task_id":    task.ID.String(),
		})
		analyticsEvents = append(analyticsEvents, &TaskAnalytics{
			ID:        uuid.New(),
			TaskID:    task.ID,
			UserID:    callerID,
			Action:    "status_changed",
			Timestamp: time.Now(),
			Metadata:  metadata,
		})
	}
	if input.Priority != nil && *input.Priority != task.Priority {
		task.Priority = *input.Priority
		changed = true
	}
	if input.AssigneeID != nil && (oldAssignee == nil || *input.AssigneeID != *oldAssignee) {
		task.AssigneeID = input.AssigneeID
		changed = true
		metadata := marshalTaskMetadata(map[string]interface{}{
			"old_assignee": func() string {
				if oldAssignee != nil {
					return oldAssignee.String()
				} else {
					return ""
				}
			}(),
			"new_assignee": input.AssigneeID.String(),
			"updated_by":   callerID.String(),
			"task_id":      task.ID.String(),
		})
		analyticsEvents = append(analyticsEvents, &TaskAnalytics{
			ID:        uuid.New(),
			TaskID:    task.ID,
			UserID:    callerID,
			Action:    "assignee_changed",
			Timestamp: time.Now(),
			Metadata:  metadata,
		})
	}
	if input.Dependencies != nil && !equalUUIDSlices(input.Dependencies, oldDependencies) {
		task.Dependencies = input.Dependencies
		changed = true
		metadata := marshalTaskMetadata(map[string]interface{}{
			"old_dependencies": oldDependencies,
			"new_dependencies": input.Dependencies,
			"updated_by":       callerID.String(),
			"task_id":          task.ID.String(),
		})
		analyticsEvents = append(analyticsEvents, &TaskAnalytics{
			ID:        uuid.New(),
			TaskID:    task.ID,
			UserID:    callerID,
			Action:    "dependencies_changed",
			Timestamp: time.Now(),
			Metadata:  metadata,
		})
	}
	// ... handle other fields as needed ...

	task.UpdatedAt = time.Now()
	err = s.repo.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	if changed {
		for _, event := range analyticsEvents {
			_ = s.repo.RecordTaskActivity(ctx, event)
		}
	} else {
		// No significant field changed, but still record a generic update event
		metadata := marshalTaskMetadata(map[string]interface{}{
			"updated_by": callerID.String(),
			"task_id":    task.ID.String(),
		})
		analytics := &TaskAnalytics{
			ID:        uuid.New(),
			TaskID:    task.ID,
			UserID:    callerID,
			Action:    "task_updated",
			Timestamp: time.Now(),
			Metadata:  metadata,
		}
		_ = s.repo.RecordTaskActivity(ctx, analytics)
	}

	s.recordTaskActivity(ctx, task, task.CreatorID, "task_updated", map[string]interface{}{
		"title":  task.Title,
		"status": task.Status,
	})

	return task, nil
}

// Helper to compare slices of UUIDs
func equalUUIDSlices(a, b []uuid.UUID) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[uuid.UUID]int)
	for _, v := range a {
		m[v]++
	}
	for _, v := range b {
		if m[v] == 0 {
			return false
		}
		m[v]--
	}
	return true
}

func (s *service) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status TaskStatus) (*Task, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	if !status.IsValid() {
		return nil, ErrInvalidInput
	}

	if !isValidStatusTransition(task.Status, status) {
		return nil, ErrInvalidTransition
	}

	// Check dependencies if moving to completed
	if status == TaskStatusCompleted {
		completed, err := s.checkDependenciesCompleted(ctx, task.Dependencies)
		if err != nil {
			return nil, err
		}
		if !completed {
			return nil, ErrDependencyFailed
		}
	}

	oldStatus := task.Status
	task.Status = status
	task.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	// Record status change activity
	if callerID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		metadata := marshalTaskMetadata(map[string]interface{}{
			"old_status": string(oldStatus),
			"new_status": string(status),
			"updated_by": callerID.String(),
			"task_id":    task.ID.String(),
		})
		analytics := &TaskAnalytics{
			ID:        uuid.New(),
			TaskID:    task.ID,
			UserID:    callerID,
			Action:    "status_changed",
			Timestamp: time.Now(),
			Metadata:  metadata,
		}
		_ = s.repo.RecordTaskActivity(ctx, analytics)
	}

	// Publish dashboard event
	event := &events.DashboardEvent{
		EventType: events.DashboardEventCacheInvalidate,
		UserID:    task.CreatorID,
		Timestamp: time.Now().UTC(),
		Details: map[string]interface{}{
			"action":  "task_status_updated",
			"task_id": id,
			"status":  status,
		},
	}
	if err := s.redis.PublishDashboardEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish dashboard event", zap.Error(err))
	}

	s.recordTaskActivity(ctx, task, task.CreatorID, "status_changed", map[string]interface{}{
		"old_status": string(oldStatus),
		"new_status": string(status),
	})

	return task, nil
}

func (s *service) DeleteTask(ctx context.Context, id uuid.UUID) error {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrTaskNotFound
	}

	// Record deletion activity before deleting
	if callerID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		s.recordTaskDeletion(ctx, task.ID, callerID)
	}

	s.recordTaskActivity(ctx, task, task.CreatorID, "task_deleted", map[string]interface{}{
		"title":  task.Title,
		"status": task.Status,
	})

	// Publish dashboard event
	event := &events.DashboardEvent{
		EventType: events.DashboardEventCacheInvalidate,
		UserID:    task.CreatorID,
		Timestamp: time.Now().UTC(),
		Details: map[string]interface{}{
			"action":  "task_deleted",
			"task_id": id,
		},
	}
	if err := s.redis.PublishDashboardEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish dashboard event", zap.Error(err))
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) recordTaskDeletion(ctx context.Context, taskID, userID uuid.UUID) {
	analytics := &TaskAnalytics{
		ID:        uuid.New(),
		TaskID:    taskID,
		UserID:    userID,
		Action:    "task_deleted",
		Timestamp: time.Now(),
	}

	_ = s.repo.RecordTaskActivity(ctx, analytics)
}

func (s *service) GetTaskMetrics(ctx context.Context, id uuid.UUID) (*TaskMetrics, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	metrics := &TaskMetrics{
		HealthScore:     calculateHealthScore(task),
		ComplexityScore: calculateComplexityScore(task),
		ProgressMetrics: task.ProgressMetrics,
		Blockers:        task.Blockers,
		RiskFactors:     task.RiskFactors,
	}

	return metrics, nil
}

func (s *service) checkDependenciesCompleted(ctx context.Context, dependencies []uuid.UUID) (bool, error) {
	for _, depID := range dependencies {
		dep, err := s.repo.FindByID(ctx, depID)
		if err != nil {
			return false, err
		}
		if dep == nil || dep.Status != TaskStatusCompleted {
			return false, nil
		}
	}
	return true, nil
}

func isValidStatusTransition(current, new TaskStatus) bool {
	transitions := map[TaskStatus][]TaskStatus{
		TaskStatusUpcoming: {
			TaskStatusInProgress,
			TaskStatusCancelled,
			TaskStatusDeferred,
			TaskStatusCompleted,
		},
		TaskStatusInProgress: {
			TaskStatusCompleted,
			TaskStatusBlocked,
			TaskStatusUnderReview,
			TaskStatusUpcoming,
		},
		TaskStatusCompleted: {
			TaskStatusUpcoming,
			TaskStatusInProgress,
		},
		TaskStatusCancelled: {
			TaskStatusUpcoming,
		},
		TaskStatusBlocked: {
			TaskStatusInProgress,
			TaskStatusUpcoming,
			TaskStatusCancelled,
		},
		TaskStatusUnderReview: {
			TaskStatusCompleted,
			TaskStatusInProgress,
		},
		TaskStatusDeferred: {
			TaskStatusUpcoming,
			TaskStatusCancelled,
		},
	}

	allowed, exists := transitions[current]
	if !exists {
		return false
	}

	for _, status := range allowed {
		if status == new {
			return true
		}
	}
	return false
}

func calculateHealthScore(task *Task) float64 {
	score := 1.0

	// Status impact
	if task.Status == TaskStatusBlocked {
		score *= 0.5
	} else if task.Status == TaskStatusDeferred {
		score *= 0.7
	}

	// Due date impact
	if task.DueDate != nil && task.DueDate.Before(time.Now()) {
		score *= 0.8
	}

	// Blockers impact
	if len(task.Blockers) > 0 {
		score *= 0.9
	}

	return score
}

func calculateComplexityScore(task *Task) float64 {
	score := 1.0

	// Base complexity from estimated hours
	if task.EstimatedHours > 0 {
		score *= (1 + task.EstimatedHours/40) // 40 hours as baseline
	}

	// Dependencies impact
	if len(task.Dependencies) > 0 {
		score *= (1 + float64(len(task.Dependencies))*0.1)
	}

	// Blockers impact
	if len(task.Blockers) > 0 {
		score *= (1 + float64(len(task.Blockers))*0.2)
	}

	return score
}

func (s *service) GetProjectTasks(ctx context.Context, projectID uuid.UUID, filter TaskFilter) ([]Task, int64, error) {
	filter.ProjectID = &projectID
	return s.repo.FindAll(ctx, filter)
}

func (s *service) AssignTask(ctx context.Context, id uuid.UUID, assigneeID uuid.UUID) (*Task, error) {
	task, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	// Check if the assignee is changing
	oldAssigneeID := task.AssigneeID
	var wasAssigned bool
	var oldAssignee uuid.UUID
	if oldAssigneeID != nil {
		wasAssigned = true
		oldAssignee = *oldAssigneeID
	}

	task.AssigneeID = &assigneeID
	task.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, task)
	if err != nil {
		return nil, err
	}

	// Record assignment activity
	if callerID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		metadata := map[string]interface{}{
			"new_assignee_id": assigneeID.String(),
		}

		if wasAssigned {
			metadata["old_assignee_id"] = oldAssignee.String()
		} else {
			metadata["old_assignee_id"] = nil
		}

		s.recordTaskAssignment(ctx, task.ID, callerID, metadata)
	}

	s.recordTaskActivity(ctx, task, task.CreatorID, "task_assigned", map[string]interface{}{
		"new_assignee_id": assigneeID.String(),
	})

	return task, nil
}

func (s *service) recordTaskAssignment(ctx context.Context, taskID, userID uuid.UUID, metadata map[string]interface{}) {
	metadataJSON, _ := json.Marshal(metadata)

	analytics := &TaskAnalytics{
		ID:        uuid.New(),
		TaskID:    taskID,
		UserID:    userID,
		Action:    "task_assigned",
		Timestamp: time.Now(),
		Metadata:  string(metadataJSON),
	}

	_ = s.repo.RecordTaskActivity(ctx, analytics)
}

// Analytics implementation
func (s *service) RecordTaskActivity(ctx context.Context, input RecordTaskActivityInput) error {
	timestamp := input.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	metadata := ""
	if input.Metadata != nil {
		metadataJSON, err := json.Marshal(input.Metadata)
		if err == nil {
			metadata = string(metadataJSON)
		}
	}

	analytics := &TaskAnalytics{
		ID:        uuid.New(),
		TaskID:    input.TaskID,
		UserID:    input.UserID,
		Action:    input.Action,
		Timestamp: timestamp,
		Metadata:  metadata,
	}

	return s.repo.RecordTaskActivity(ctx, analytics)
}

func (s *service) GetTaskAnalytics(ctx context.Context, taskID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]TaskAnalytics, int64, error) {
	filter := AnalyticsFilter{
		TaskID:    &taskID,
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      page,
		PageSize:  pageSize,
	}

	return s.repo.GetTaskAnalytics(ctx, filter)
}

func (s *service) GetUserTaskAnalytics(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]TaskAnalytics, int64, error) {
	filter := AnalyticsFilter{
		UserID:    &userID,
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      page,
		PageSize:  pageSize,
	}

	return s.repo.GetTaskAnalytics(ctx, filter)
}

func (s *service) GetTaskActivitySummary(ctx context.Context, taskID uuid.UUID, startTime, endTime time.Time) (*TaskActivitySummary, error) {
	actionCounts, err := s.repo.GetTaskActivitySummary(ctx, taskID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Calculate total actions
	totalActions := 0
	for _, count := range actionCounts {
		totalActions += count
	}

	return &TaskActivitySummary{
		TaskID:       taskID,
		ActionCounts: actionCounts,
		StartTime:    startTime,
		EndTime:      endTime,
		TotalActions: totalActions,
	}, nil
}

func (s *service) GetUserTaskActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (*UserTaskActivitySummary, error) {
	actionCounts, err := s.repo.GetUserTaskActivitySummary(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Calculate total actions
	totalActions := 0
	for _, count := range actionCounts {
		totalActions += count
	}

	return &UserTaskActivitySummary{
		UserID:       userID,
		ActionCounts: actionCounts,
		StartTime:    startTime,
		EndTime:      endTime,
		TotalActions: totalActions,
	}, nil
}

func (s *service) GetDashboardMetrics(userID uuid.UUID) (TasksDashboardMetrics, error) {
	ctx := context.Background()
	filter := TaskFilter{AssigneeID: &userID}
	tasks, _, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return TasksDashboardMetrics{}, err
	}
	total := len(tasks)
	completed := 0
	overdue := 0
	now := time.Now()
	for _, t := range tasks {
		if t.Status == TaskStatusCompleted {
			completed++
		}
		if t.DueDate != nil && t.DueDate.Before(now) && t.Status != TaskStatusCompleted {
			overdue++
		}
	}
	return TasksDashboardMetrics{
		Total:     total,
		Completed: completed,
		Overdue:   overdue,
	}, nil
}

func (s *service) GetTodayTasks(ctx context.Context, userID uuid.UUID) ([]Task, error) {
	// Modified to return all active tasks for the user

	// Get all tasks assigned to the user that are not completed
	filter := TaskFilter{
		AssigneeID: &userID,
		Status:     taskStatusPtr(TaskStatusInProgress), // Only get tasks in progress
	}

	tasks, _, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Log for diagnostic purposes
	nilDueDateCount := 0
	for _, task := range tasks {
		if task.DueDate == nil {
			nilDueDateCount++
		}
	}
	s.logger.Info("GetTodayTasks results",
		zap.String("user_id", userID.String()),
		zap.Int("total_found", len(tasks)),
		zap.Int("nil_due_date_count", nilDueDateCount))

	// Return all tasks, even those with nil DueDate
	s.logger.Info("GetTodayTasks returning all active tasks",
		zap.Int("active_count", len(tasks)))

	return tasks, nil
}

// Helper function to create a TaskStatus pointer
func taskStatusPtr(status TaskStatus) *TaskStatus {
	return &status
}

func (s *service) recordTaskActivity(ctx context.Context, task *Task, userID uuid.UUID, action string, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["action"] = action

	// Publish dashboard event for cache invalidation
	event := &events.DashboardEvent{
		EventType: events.DashboardEventCacheInvalidate,
		UserID:    userID,
		EntityID:  task.ID,
		Timestamp: time.Now().UTC(),
		Details:   metadata,
	}
	if err := s.redis.PublishDashboardEvent(ctx, event); err != nil {
		s.logger.Error("Failed to publish dashboard event", zap.Error(err))
	}
}
