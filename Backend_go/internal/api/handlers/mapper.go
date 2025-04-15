package handlers

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/habits"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/task"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/todos"
)

// Habits
func HabitToResponse(h *habits.Habit) *dto.HabitResponse {
	if h == nil {
		return nil
	}
	return &dto.HabitResponse{
		ID:                h.ID,
		UserID:            h.UserID,
		Title:             h.Title,
		Description:       h.Description,
		StartDay:          h.StartDay,
		EndDay:            h.EndDay,
		CurrentStreak:     h.CurrentStreak,
		StreakStartDate:   h.StreakStartDate,
		LongestStreak:     h.LongestStreak,
		IsCompleted:       h.IsCompleted,
		LastCompletedDate: h.LastCompletedDate,
		CreatedAt:         h.CreatedAt,
		UpdatedAt:         h.UpdatedAt,
		StreakQuality:     h.StreakQuality,
	}
}

func StreakHistoryToResponse(h *habits.StreakHistory) *dto.StreakHistoryResponse {
	if h == nil {
		return nil
	}
	return &dto.StreakHistoryResponse{
		ID:            h.ID,
		HabitID:       h.HabitID,
		StartDate:     h.StartDate,
		EndDate:       h.EndDate,
		StreakLength:  h.StreakLength,
		CompletedDays: h.CompletedDays,
		CreatedAt:     h.CreatedAt,
	}
}

// Tasks
func TaskToResponse(t *task.Task) *dto.TaskResponse {
	if t == nil {
		return nil
	}
	return &dto.TaskResponse{
		ID:             t.ID,
		Title:          t.Title,
		Description:    t.Description,
		Status:         string(t.Status),
		Priority:       string(t.Priority),
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		CreatorID:      t.CreatorID,
		AssigneeID:     t.AssigneeID,
		ReviewerID:     t.ReviewerID,
		CategoryID:     t.CategoryID,
		ParentTaskID:   t.ParentTaskID,
		ProjectID:      t.ProjectID,
		OrganizationID: t.OrganizationID,
		EstimatedHours: t.EstimatedHours,
		StartDate:      t.StartDate,
		Duration:       t.Duration,
		DueDate:        t.DueDate,
	}
}

func TasksToResponse(tasks []task.Task) []*dto.TaskResponse {
	response := make([]*dto.TaskResponse, len(tasks))
	for i, t := range tasks {
		response[i] = TaskToResponse(&t)
	}
	return response
}

// Todos
func TodoToResponse(t *todos.Todo) *dto.TodoResponse {
	if t == nil {
		return nil
	}
	return &dto.TodoResponse{
		ID:                    t.ID,
		Title:                 t.Title,
		Description:           t.Description,
		Status:                string(t.Status),
		Priority:              string(t.Priority),
		DueDate:               t.DueDate,
		ReminderTime:          t.ReminderTime,
		IsRecurring:           t.IsRecurring,
		RecurrencePattern:     t.RecurrencePattern,
		Tags:                  t.Tags,
		Checklist:             t.Checklist,
		LinkedTaskID:          t.LinkedTaskID,
		LinkedCalendarEventID: t.LinkedCalendarEventID,
		IsCompleted:           t.IsCompleted,
		CompletedAt:           t.CompletionDate,
		CreatedAt:             t.CreatedAt,
		UpdatedAt:             t.UpdatedAt,
		UserID:                t.UserID,
		ListID:                t.ListID,
	}
}

func TodosToResponse(todos []todos.Todo) []*dto.TodoResponse {
	response := make([]*dto.TodoResponse, len(todos))
	for i, t := range todos {
		response[i] = TodoToResponse(&t)
	}
	return response
}

func TodoListToResponse(l *todos.TodoList) *dto.TodoListResponse {
	if l == nil {
		return nil
	}
	return &dto.TodoListResponse{
		ID:          l.ID,
		Name:        l.Name,
		Description: l.Description,
		IsDefault:   l.IsDefault,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
		UserID:      l.UserID,
		Todos:       TodosToResponse(l.Todos),
		TotalCount:  int64(len(l.Todos)),
		Page:        1,
		PageSize:    20,
	}
}

func TodoListsToResponse(lists []todos.TodoList) []*dto.TodoListResponse {
	response := make([]*dto.TodoListResponse, len(lists))
	for i, l := range lists {
		response[i] = TodoListToResponse(&l)
	}
	return response
}
