package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/calendar"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/events"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/habits"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/task"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/todos"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/user"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/cache"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DashboardHandler struct {
	habitsService   habits.Service
	tasksService    task.Service
	todosService    todos.Service
	calendarService calendar.Service
	userService     user.Service
	redisClient     *cache.RedisClient
	logger          *zap.Logger
}

func NewDashboardHandler(
	habitsService habits.Service,
	tasksService task.Service,
	todosService todos.Service,
	calendarService calendar.Service,
	userService user.Service,
	redisClient *cache.RedisClient,
	logger *zap.Logger,
) *DashboardHandler {
	return &DashboardHandler{
		habitsService:   habitsService,
		tasksService:    tasksService,
		todosService:    todosService,
		calendarService: calendarService,
		userService:     userService,
		redisClient:     redisClient,
		logger:          logger,
	}
}

// Conversion functions from domain metrics to DTO metrics
func HabitsDashboardMetricsToDTO(m habits.HabitsDashboardMetrics) dto.HabitsDashboardMetrics {
	return dto.HabitsDashboardMetrics{
		Total:     m.Total,
		Active:    m.Active,
		Completed: m.Completed,
		Streak:    m.Streak,
	}
}

func TasksDashboardMetricsToDTO(m task.TasksDashboardMetrics) dto.TasksDashboardMetrics {
	return dto.TasksDashboardMetrics{
		Total:     m.Total,
		Completed: m.Completed,
		Overdue:   m.Overdue,
	}
}

func TodosDashboardMetricsToDTO(m todos.TodosDashboardMetrics) dto.TodosDashboardMetrics {
	return dto.TodosDashboardMetrics{
		Total:     m.Total,
		Completed: m.Completed,
		Overdue:   m.Overdue,
	}
}

func CalendarDashboardMetricsToDTO(m calendar.CalendarDashboardMetrics) dto.CalendarDashboardMetrics {
	return dto.CalendarDashboardMetrics{
		Upcoming: m.Upcoming,
		Total:    m.Total,
	}
}

func UserDashboardMetricsToDTO(m user.UserDashboardMetrics) dto.UserDashboardMetrics {
	return dto.UserDashboardMetrics{
		ActivitySummary: m.ActivitySummary,
	}
}

func (h *DashboardHandler) GetDashboardMetrics(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Use the standardized cache key
	cacheKey := fmt.Sprintf("dashboard:metrics:%v", userID)
	cachedData, err := h.redisClient.Get(c.Request.Context(), cacheKey)
	if err == nil && cachedData != "" {
		var response dto.DashboardMetricsResponse
		if unmarshalErr := json.Unmarshal([]byte(cachedData), &response); unmarshalErr == nil {
			c.JSON(http.StatusOK, gin.H{"data": response})
			return
		}
	}

	// Collect metrics from all services
	habitsMetrics, err := h.habitsService.GetDashboardMetrics(userID)
	if err != nil {
		h.logger.Error("Failed to get habits metrics", zap.Error(err))
	}

	tasksMetrics, err := h.tasksService.GetDashboardMetrics(userID)
	if err != nil {
		h.logger.Error("Failed to get tasks metrics", zap.Error(err))
	}

	todosMetrics, err := h.todosService.GetDashboardMetrics(userID)
	if err != nil {
		h.logger.Error("Failed to get todos metrics", zap.Error(err))
	}

	calendarMetrics, err := h.calendarService.GetDashboardMetrics(userID)
	if err != nil {
		h.logger.Error("Failed to get calendar metrics", zap.Error(err))
	}

	userMetrics, err := h.userService.GetDashboardMetrics(userID)
	if err != nil {
		h.logger.Error("Failed to get user metrics", zap.Error(err))
	}

	// Get habit heatmap data (default to month period)
	habitHeatmap, err := h.habitsService.GetHeatmapData(c.Request.Context(), userID, "month")
	if err != nil {
		h.logger.Error("Failed to get habit heatmap data", zap.Error(err))
		// Initialize with empty map if there's an error
		habitHeatmap = make(map[string]int)
	}

	// Collect today's items for the timeline
	var timeline []dto.TimelineItem

	// Track counts for each type
	habitCount := 0
	taskCount := 0
	todoCount := 0
	eventCount := 0

	todayHabits, err := h.habitsService.GetHabitsDueToday(c.Request.Context(), userID)
	if err == nil {
		for _, habit := range todayHabits {
			startTime := habit.StartDay
			// Use current time for habits without specific time
			if startTime.IsZero() {
				startTime = time.Now()
			}
			timeline = append(timeline, dto.TimelineItem{
				ID:          habit.ID,
				Title:       habit.Title,
				StartTime:   startTime,
				Type:        "habit",
				IsCompleted: habit.IsCompleted,
			})
			habitCount++
		}
	} else {
		h.logger.Error("Failed to get habits due today", zap.Error(err))
	}

	todayTasks, err := h.tasksService.GetTodayTasks(c.Request.Context(), userID)
	if err == nil {
		for _, t := range todayTasks {
			startTime := t.StartDate
			endTime := t.DueDate

			// If due date is nil, create a reasonable default end time
			if endTime == nil {
				defaultEndTime := startTime.Add(1 * time.Hour)
				endTime = &defaultEndTime
			}

			timeline = append(timeline, dto.TimelineItem{
				ID:          t.ID,
				Title:       t.Title,
				StartTime:   startTime,
				EndTime:     endTime,
				Type:        "task",
				IsCompleted: t.Status == "Completed",
			})
			taskCount++
		}
	} else {
		h.logger.Error("Failed to get tasks due today", zap.Error(err))
	}

	todayTodos, err := h.todosService.GetTodayTodos(c.Request.Context(), userID)
	if err == nil {
		for _, todo := range todayTodos {
			// If todo doesn't have a due date, use current time as default
			startTime := time.Now()
			if todo.DueDate != nil {
				startTime = *todo.DueDate
			}

			timeline = append(timeline, dto.TimelineItem{
				ID:          todo.ID,
				Title:       todo.Title,
				StartTime:   startTime,
				Type:        "todo",
				IsCompleted: todo.IsCompleted,
			})
			todoCount++
		}
	} else {
		h.logger.Error("Failed to get todos due today", zap.Error(err))
	}

	todayEvents, err := h.calendarService.GetTodayEvents(c.Request.Context(), userID)
	if err == nil {
		for _, event := range todayEvents {
			timeline = append(timeline, dto.TimelineItem{
				ID:        event.ID,
				Title:     event.Title,
				StartTime: event.StartTime,
				EndTime:   &event.EndTime,
				Type:      "event",
			})
			eventCount++
		}
	} else {
		h.logger.Error("Failed to get events due today", zap.Error(err))
	}

	// Get upcoming events for the calendar widget (max 4)
	upcomingEvents, err := h.calendarService.GetUpcomingEvents(c.Request.Context(), userID, 4)
	if err == nil {
		// Add a special field to differentiate from today's events
		for _, event := range upcomingEvents {
			// Skip events that are already in the timeline (from today)
			alreadyInTimeline := false
			for _, item := range timeline {
				if item.ID == event.ID && item.Type == "event" {
					alreadyInTimeline = true
					break
				}
			}

			if !alreadyInTimeline {
				timeline = append(timeline, dto.TimelineItem{
					ID:        event.ID,
					Title:     event.Title,
					StartTime: event.StartTime,
					EndTime:   &event.EndTime,
					Type:      "event",
				})
			}
		}
	} else {
		h.logger.Error("Failed to get upcoming events", zap.Error(err))
	}

	h.logger.Info("Timeline items collected",
		zap.String("user_id", userID.String()),
		zap.Int("habit_count", habitCount),
		zap.Int("task_count", taskCount),
		zap.Int("todo_count", todoCount),
		zap.Int("event_count", eventCount),
		zap.Int("total_timeline_items", len(timeline)),
		zap.Bool("timeline", len(timeline) > 0))

	response := dto.DashboardMetricsResponse{
		Habits:        HabitsDashboardMetricsToDTO(habitsMetrics),
		Tasks:         TasksDashboardMetricsToDTO(tasksMetrics),
		Todos:         TodosDashboardMetricsToDTO(todosMetrics),
		Calendar:      CalendarDashboardMetricsToDTO(calendarMetrics),
		User:          UserDashboardMetricsToDTO(userMetrics),
		DailyTimeline: timeline,
		HabitHeatmap:  habitHeatmap,
		Timestamp:     time.Now().UTC(),
	}

	// Cache the response using the new key
	if data, err := json.Marshal(response); err == nil {
		if err := h.redisClient.Set(c.Request.Context(), cacheKey, string(data), 5*time.Minute); err != nil {
			h.logger.Error("Failed to cache dashboard metrics", zap.Error(err))
		}

		// Publish a dashboard event to notify other services about the updated metrics
		dashboardEvent := &events.DashboardEvent{
			EventType: events.DashboardEventMetricsUpdate,
			UserID:    userID,
			Timestamp: time.Now().UTC(),
			Details: map[string]interface{}{
				"source": "go_backend",
			},
		}

		if err := h.redisClient.PublishDashboardEvent(c.Request.Context(), dashboardEvent); err != nil {
			h.logger.Error("Failed to publish dashboard metrics update event", zap.Error(err))
		} else {
			h.logger.Info("Published dashboard metrics update event", zap.String("user_id", userID.String()))
		}
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// StartDashboardEventListener starts listening for dashboard events
func (h *DashboardHandler) StartDashboardEventListener(ctx context.Context) {
	go func() {
		err := h.redisClient.SubscribeToDashboardEvents(ctx, func(event *events.DashboardEvent) error {
			h.logger.Info("Received dashboard event",
				zap.String("event_type", event.EventType),
				zap.String("user_id", event.UserID.String()),
				zap.Any("details", event.Details))

			// Invalidate both possible dashboard cache key patterns for the affected user
			patterns := []string{
				fmt.Sprintf("compass:dashboard:*:%s", event.UserID.String()),
				fmt.Sprintf("dashboard:metrics:%s", event.UserID.String()),
			}
			for _, pattern := range patterns {
				if err := h.redisClient.ClearByPattern(ctx, pattern); err != nil {
					h.logger.Error("Failed to invalidate dashboard cache",
						zap.Error(err),
						zap.String("pattern", pattern))
				} else {
					h.logger.Info("Successfully invalidated dashboard cache",
						zap.String("user_id", event.UserID.String()),
						zap.String("pattern", pattern))
				}
			}
			return nil
		})
		if err != nil {
			h.logger.Error("Dashboard event listener error", zap.Error(err))
		}
	}()
}
