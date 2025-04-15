package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/habits"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HabitsHandler handles HTTP requests for habits operations
type HabitsHandler struct {
	service habits.Service
}

// NewHabitsHandler creates a new HabitsHandler instance
func NewHabitsHandler(service habits.Service) *HabitsHandler {
	return &HabitsHandler{service: service}
}

// CreateHabit godoc
// @Summary Create a new habit
// @Description Create a new habit with the provided information
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param habit body dto.CreateHabitRequest true "Habit creation request"
// @Success 201 {object} dto.HabitResponse "Habit created successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits [post]
func (h *HabitsHandler) CreateHabit(c *gin.Context) {
	// Get validated model from context (set by validation middleware)
	var req dto.CreateHabitRequest
	validatedModel, exists := c.Get("validated_model")

	if exists {
		// If validation middleware provided the model, use it
		// The model will be a pointer since we created it with reflect.New
		if validatedPtr, ok := validatedModel.(*dto.CreateHabitRequest); ok {
			req = *validatedPtr
		} else {
			// Log the actual type for debugging
			log.Errorf("Invalid model type: %T, expected *dto.CreateHabitRequest", validatedModel)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model type from validation"})
			return
		}
	} else {
		// If validation middleware didn't run, do manual binding
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	input := habits.CreateHabitInput{
		Title:       req.Title,
		Description: req.Description,
		StartDay:    req.StartDay,
		EndDay:      req.EndDay,
		UserID:      userID,
	}

	createdHabit, err := h.service.CreateHabit(c.Request.Context(), input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == habits.ErrInvalidInput {
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": HabitToResponse(createdHabit)})
}

// GetHabit godoc
// @Summary Get a habit by ID
// @Description Get detailed information about a specific habit
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID" format(uuid)
// @Success 200 {object} dto.HabitResponse "Habit details retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid habit ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Habit not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/{id} [get]
func (h *HabitsHandler) GetHabit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	habit, err := h.service.GetHabit(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == habits.ErrHabitNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// Record habit view analytics
	userID, exists := middleware.GetUserID(c)
	if exists {
		go func() {
			ctx := context.Background()
			h.service.RecordHabitActivity(ctx, habits.RecordHabitActivityInput{
				HabitID: id,
				UserID:  userID,
				Action:  habits.ActionHabitView,
				Metadata: map[string]interface{}{
					"title":  habit.Title,
					"via":    "api",
					"path":   c.Request.URL.Path,
					"method": c.Request.Method,
				},
			})
		}()
	}

	// Explicitly set content type (must change it in the future)
	c.Header("Content-Type", "application/json; charset=utf-8")

	c.JSON(http.StatusOK, gin.H{"data": HabitToResponse(habit)})
}

// ListHabits godoc
// @Description Get a list of all habits for the authenticated user
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of habits per page" default(10)
// @Success 200 {object} dto.HabitListResponse "List of habits retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits [get]
func (h *HabitsHandler) ListHabits(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	pageStr := c.DefaultQuery("page", "0")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page size"})
		return
	}

	filter := habits.HabitFilter{
		Page:     page,
		PageSize: pageSize,
		UserID:   &userID,
	}

	habitsData, total, err := h.service.ListHabits(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Record habit list view analytics
	go func() {
		ctx := context.Background()
		h.service.RecordHabitActivity(ctx, habits.RecordHabitActivityInput{
			HabitID: uuid.Nil, // No specific habit ID for list view
			UserID:  userID,
			Action:  habits.ActionHabitListView,
			Metadata: map[string]interface{}{
				"page":      page,
				"page_size": pageSize,
				"total":     total,
				"count":     len(habitsData),
				"via":       "api",
				"path":      c.Request.URL.Path,
				"method":    c.Request.Method,
			},
		})
	}()

	responses := make([]dto.HabitResponse, len(habitsData))
	for i, habit := range habitsData {
		response := HabitToResponse(&habit)
		responses[i] = *response
	}

	// Explicitly set content type
	c.Header("Content-Type", "application/json; charset=utf-8")

	c.JSON(http.StatusOK, gin.H{"data": dto.HabitListResponse{
		Habits:     responses,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}})
}

// UpdateHabit godoc
// @Summary Update a habit
// @Description Update an existing habit's information
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID" format(uuid)
// @Param habit body dto.UpdateHabitRequest true "Habit update information"
// @Success 200 {object} dto.HabitResponse "Habit updated successfully"
// @Failure 400 {object} map[string]string "Invalid request or habit ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Habit not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/{id} [put]
func (h *HabitsHandler) UpdateHabit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	var req dto.UpdateHabitRequest
	validatedModel, exists := c.Get("validated_model")

	if exists {
		// If validation middleware provided the model, use it
		if validatedPtr, ok := validatedModel.(*dto.UpdateHabitRequest); ok {
			req = *validatedPtr
		} else {
			log.Errorf("Invalid model type: %T, expected *dto.UpdateHabitRequest", validatedModel)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model type from validation"})
			return
		}
	} else {
		// If validation middleware didn't run, do manual binding
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	input := habits.UpdateHabitInput{
		Title:       req.Title,
		Description: req.Description,
		StartDay:    req.StartDay,
		EndDay:      req.EndDay,
	}

	updatedHabit, err := h.service.UpdateHabit(c.Request.Context(), id, input)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == habits.ErrHabitNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": HabitToResponse(updatedHabit)})
}

// DeleteHabit godoc
// @Summary Delete a habit
// @Description Delete a habit by ID
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID" format(uuid)
// @Success 204 "Habit deleted successfully"
// @Failure 400 {object} map[string]string "Invalid habit ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Habit not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/{id} [delete]
func (h *HabitsHandler) DeleteHabit(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	err = h.service.DeleteHabit(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == habits.ErrHabitNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetStreakHistory godoc
// @Summary Get streak history for a habit
// @Description Get the streak history for a specific habit
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID" format(uuid)
// @Success 200 {array} dto.StreakHistoryResponse "Streak history retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid habit ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Habit not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/{id}/streak-history [get]
func (h *HabitsHandler) GetStreakHistory(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	history, err := h.service.GetStreakHistory(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == habits.ErrHabitNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// Record streak history view analytics
	userID, exists := middleware.GetUserID(c)
	if exists {
		go func() {
			ctx := context.Background()
			h.service.RecordHabitActivity(ctx, habits.RecordHabitActivityInput{
				HabitID: id,
				UserID:  userID,
				Action:  habits.ActionStreakHistoryView,
				Metadata: map[string]interface{}{
					"history_count": len(history),
					"via":           "api",
					"path":          c.Request.URL.Path,
					"method":        c.Request.Method,
				},
			})
		}()
	}

	responses := make([]dto.StreakHistoryResponse, len(history))
	for i, h := range history {
		responses[i] = *StreakHistoryToResponse(&h)
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

// GetHabitsDueToday godoc
// @Summary Get habits due today
// @Description Get all habits that are due for completion today
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.HabitResponse "Habits due today retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/due-today [get]
func (h *HabitsHandler) GetHabitsDueToday(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	habitsData, err := h.service.GetHabitsDueToday(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Record habits due today view analytics
	go func() {
		ctx := context.Background()
		h.service.RecordHabitActivity(ctx, habits.RecordHabitActivityInput{
			HabitID: uuid.Nil, // No specific habit ID
			UserID:  userID.(uuid.UUID),
			Action:  habits.ActionHabitDueTodayView,
			Metadata: map[string]interface{}{
				"count":  len(habitsData),
				"via":    "api",
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
			},
		})
	}()

	responses := make([]dto.HabitResponse, len(habitsData))
	for i, habit := range habitsData {
		responses[i] = *HabitToResponse(&habit)
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}

// GetHabitStats godoc
// @Summary Get habit statistics
// @Description Get statistics for a specific habit
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID" format(uuid)
// @Success 200 {object} dto.HabitStatsResponse "Habit statistics retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid habit ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Habit not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/{id}/stats [get]
func (h *HabitsHandler) GetHabitStats(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	habit, err := h.service.GetHabit(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == habits.ErrHabitNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// Record habit stats view analytics
	userID, exists := middleware.GetUserID(c)
	if exists {
		go func() {
			ctx := context.Background()
			h.service.RecordHabitActivity(ctx, habits.RecordHabitActivityInput{
				HabitID: id,
				UserID:  userID,
				Action:  habits.ActionHabitStats,
				Metadata: map[string]interface{}{
					"title":          habit.Title,
					"current_streak": habit.CurrentStreak,
					"longest_streak": habit.LongestStreak,
					"streak_quality": habit.StreakQuality,
					"via":            "api",
					"path":           c.Request.URL.Path,
					"method":         c.Request.Method,
				},
			})
		}()
	}

	stats := dto.HabitStatsResponse{
		TotalHabits:     1,
		ActiveHabits:    1,
		CompletedHabits: 0,
	}

	if habit.IsCompleted {
		stats.CompletedHabits = 1
		stats.ActiveHabits = 0
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetHabitHeatmap godoc
// @Summary Get habit completion heatmap data
// @Description Get aggregated habit completion data for visualization as a heatmap
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param period query string false "Time period for heatmap data (week, month, year)" Enums(week, month, year) default(year)
// @Success 200 {object} dto.HeatmapResponse "Heatmap data retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/heatmap [get]
func (h *HabitsHandler) GetHabitHeatmap(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Get period from query parameters, default to year if not specified
	period := c.DefaultQuery("period", "year")
	if period != "week" && period != "month" && period != "year" {
		period = "year" // Default to year for invalid values
	}

	// Get heatmap data from service
	heatmapData, err := h.service.GetHeatmapData(c.Request.Context(), userID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Record heatmap view analytics
	go func() {
		ctx := context.Background()
		h.service.RecordHabitActivity(ctx, habits.RecordHabitActivityInput{
			HabitID: uuid.Nil, // No specific habit ID
			UserID:  userID,
			Action:  habits.ActionHabitHeatmapView,
			Metadata: map[string]interface{}{
				"period":      period,
				"data_points": len(heatmapData),
				"via":         "api",
				"path":        c.Request.URL.Path,
				"method":      c.Request.Method,
			},
		})
	}()

	// Find min and max values for the heatmap scale
	minValue := 0 // Minimum is always 0 for habit completions
	maxValue := 0
	for _, count := range heatmapData {
		if count > maxValue {
			maxValue = count
		}
	}

	// Return the response
	c.JSON(http.StatusOK, gin.H{"data": dto.HeatmapResponse{
		Data:     heatmapData,
		Period:   period,
		MinValue: minValue,
		MaxValue: maxValue,
	}})
}

// GetHabitAnalytics godoc
// @Summary Get habit analytics
// @Description Get analytics data for a specific habit
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID" format(uuid)
// @Param start_time query string true "Start time (RFC3339)" format(date-time)
// @Param end_time query string true "End time (RFC3339)" format(date-time)
// @Param page query int false "Page number (default: 0)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} dto.HabitAnalyticsListResponse "Habit analytics retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Habit not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/{id}/analytics [get]
func (h *HabitsHandler) GetHabitAnalytics(c *gin.Context) {
	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	// Parse time strings to time.Time
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, expected RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, expected RFC3339"})
		return
	}

	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "0")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page size"})
		return
	}

	// Get analytics data
	analytics, total, err := h.service.GetHabitAnalytics(
		c.Request.Context(),
		habitID,
		startTime,
		endTime,
		page,
		pageSize,
	)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == habits.ErrHabitNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// Convert domain entities to DTO responses
	responseItems := make([]dto.HabitAnalyticsResponse, len(analytics))
	for i, item := range analytics {
		// Parse metadata JSON
		var metadata map[string]interface{}
		if item.Metadata != "" {
			if err := json.Unmarshal([]byte(item.Metadata), &metadata); err != nil {
				// Just use empty metadata if parsing fails
				metadata = make(map[string]interface{})
			}
		} else {
			metadata = make(map[string]interface{})
		}

		responseItems[i] = dto.HabitAnalyticsResponse{
			ID:        item.ID,
			HabitID:   item.HabitID,
			UserID:    item.UserID,
			Action:    item.Action,
			Timestamp: item.Timestamp,
			Metadata:  metadata,
		}
	}

	response := dto.HabitAnalyticsListResponse{
		Analytics:  responseItems,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetUserHabitAnalytics godoc
// @Summary Get user habit analytics
// @Description Get analytics data for habits associated with the current user
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_time query string true "Start time (RFC3339)" format(date-time)
// @Param end_time query string true "End time (RFC3339)" format(date-time)
// @Param page query int false "Page number (default: 0)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} dto.HabitAnalyticsListResponse "User habit analytics retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/analytics/user [get]
func (h *HabitsHandler) GetUserHabitAnalytics(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Parse time strings to time.Time
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, expected RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, expected RFC3339"})
		return
	}

	// Parse pagination parameters
	pageStr := c.DefaultQuery("page", "0")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page number"})
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page size"})
		return
	}

	// Get analytics data
	analytics, total, err := h.service.GetUserHabitAnalytics(
		c.Request.Context(),
		userID,
		startTime,
		endTime,
		page,
		pageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert domain entities to DTO responses
	responseItems := make([]dto.HabitAnalyticsResponse, len(analytics))
	for i, item := range analytics {
		// Parse metadata JSON
		var metadata map[string]interface{}
		if item.Metadata != "" {
			if err := json.Unmarshal([]byte(item.Metadata), &metadata); err != nil {
				// Just use empty metadata if parsing fails
				metadata = make(map[string]interface{})
			}
		} else {
			metadata = make(map[string]interface{})
		}

		responseItems[i] = dto.HabitAnalyticsResponse{
			ID:        item.ID,
			HabitID:   item.HabitID,
			UserID:    item.UserID,
			Action:    item.Action,
			Timestamp: item.Timestamp,
			Metadata:  metadata,
		}
	}

	response := dto.HabitAnalyticsListResponse{
		Analytics:  responseItems,
		TotalCount: total,
		Page:       page,
		PageSize:   pageSize,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetHabitActivitySummary godoc
// @Summary Get habit activity summary
// @Description Get a summary of activity counts by action type for a habit
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID" format(uuid)
// @Param start_time query string true "Start time (RFC3339)" format(date-time)
// @Param end_time query string true "End time (RFC3339)" format(date-time)
// @Success 200 {object} dto.HabitActivitySummaryResponse "Habit activity summary retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Habit not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/{id}/analytics/summary [get]
func (h *HabitsHandler) GetHabitActivitySummary(c *gin.Context) {
	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	// Parse time strings to time.Time
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, expected RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, expected RFC3339"})
		return
	}

	// Get summary data
	summary, err := h.service.GetHabitActivitySummary(
		c.Request.Context(),
		habitID,
		startTime,
		endTime,
	)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == habits.ErrHabitNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	response := dto.HabitActivitySummaryResponse{
		HabitID:      summary.HabitID,
		ActionCounts: summary.ActionCounts,
		StartTime:    summary.StartTime,
		EndTime:      summary.EndTime,
		TotalActions: summary.TotalActions,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetUserHabitActivitySummary godoc
// @Summary Get user habit activity summary
// @Description Get a summary of activity counts by action type for a user's habits
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_time query string true "Start time (RFC3339)" format(date-time)
// @Param end_time query string true "End time (RFC3339)" format(date-time)
// @Success 200 {object} map[string]interface{} "User habit activity summary retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/analytics/user/summary [get]
func (h *HabitsHandler) GetUserHabitActivitySummary(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Parse time strings to time.Time
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, expected RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, expected RFC3339"})
		return
	}

	// Get summary data
	summary, err := h.service.GetUserHabitActivitySummary(
		c.Request.Context(),
		userID,
		startTime,
		endTime,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := map[string]interface{}{
		"user_id":       summary.UserID,
		"action_counts": summary.ActionCounts,
		"start_time":    summary.StartTime,
		"end_time":      summary.EndTime,
		"total_actions": summary.TotalActions,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// RecordHabitActivity godoc
// @Summary Record habit activity
// @Description Manually record a habit activity for analytics
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID" format(uuid)
// @Param activity body dto.RecordHabitActivityRequest true "Activity details"
// @Success 201 "Activity recorded successfully"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Habit not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/{id}/analytics/record [post]
func (h *HabitsHandler) RecordHabitActivity(c *gin.Context) {
	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var request dto.RecordHabitActivityRequest
	validatedModel, exists := c.Get("validated_model")

	if exists {
		// If validation middleware provided the model, use it
		if validatedPtr, ok := validatedModel.(*dto.RecordHabitActivityRequest); ok {
			request = *validatedPtr
		} else {
			log.Errorf("Invalid model type: %T, expected *dto.RecordHabitActivityRequest", validatedModel)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model type from validation"})
			return
		}
	} else {
		// If validation middleware didn't run, do manual binding
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Verify habit exists
	habit, err := h.service.GetHabit(c.Request.Context(), habitID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == habits.ErrHabitNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// Check if the user owns the habit or has permission
	if habit.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "you don't have permission to record activity for this habit"})
		return
	}

	input := habits.RecordHabitActivityInput{
		HabitID:   habitID,
		UserID:    userID,
		Action:    request.Action,
		Metadata:  request.Metadata,
		Timestamp: time.Now(),
	}

	if err := h.service.RecordHabitActivity(c.Request.Context(), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// MarkHabitCompleted godoc
// @Summary Mark a habit as completed
// @Description Mark a specific habit as completed for today or a specific date
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID" format(uuid)
// @Param completion_date query string false "Completion date (YYYY-MM-DD)" format(date)
// @Success 200 {object} map[string]string "Habit marked as completed"
// @Failure 400 {object} map[string]string "Invalid habit ID or date"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Habit not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/{id}/complete [post]
func (h *HabitsHandler) MarkHabitCompleted(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Check if we have a validated model with completion date
	var completionDate *time.Time
	validatedModel, modelExists := c.Get("validated_model")

	if modelExists {
		// If validation middleware provided the model, use it
		if validatedPtr, ok := validatedModel.(*dto.HabitCompletionRequest); ok {
			completionDate = validatedPtr.CompletionDate
		} else {
			log.Errorf("Invalid model type: %T, expected *dto.HabitCompletionRequest", validatedModel)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model type from validation"})
			return
		}
	} else {
		// If validation middleware didn't run, check for completion_date in query
		if dateStr := c.Query("completion_date"); dateStr != "" {
			date, err := time.Parse("2006-01-02", dateStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid completion date format"})
				return
			}
			completionDate = &date
		}
	}

	err = h.service.MarkCompleted(c.Request.Context(), id, userID.(uuid.UUID), completionDate)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == habits.ErrHabitNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// Explicitly set content type
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.JSON(http.StatusOK, gin.H{"message": "habit marked as completed"})
}

// UnmarkHabitCompleted godoc
// @Summary Unmark a habit as completed
// @Description Remove the completion status of a habit for today
// @Tags habits
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Habit ID" format(uuid)
// @Success 200 {object} map[string]string "Habit unmarked as completed"
// @Failure 400 {object} map[string]string "Invalid habit ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "Habit not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/{id}/uncomplete [post]
func (h *HabitsHandler) UnmarkHabitCompleted(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	err = h.service.UnmarkCompleted(c.Request.Context(), id, userID.(uuid.UUID))
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == habits.ErrHabitNotFound {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "habit unmarked as completed"})
}

// GetUserHabits godoc
// @Summary Get habits by user ID
// @Description Get all habits for a specific user with optional active_only filter
// @Tags habits
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Security BearerAuth
// @Success 200 {array} dto.HabitResponse "List of user habits"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/habits/user/{user_id} [get]
func (h *HabitsHandler) GetUserHabits(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	filter := habits.HabitFilter{
		UserID:   &userID,
		Page:     0,
		PageSize: 100, // You might want to make this configurable
	}

	habitsData, _, err := h.service.ListHabits(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Record user habits view analytics
	currentUserID, exists := middleware.GetUserID(c)
	if exists {
		go func() {
			ctx := context.Background()
			h.service.RecordHabitActivity(ctx, habits.RecordHabitActivityInput{
				HabitID: uuid.Nil, // No specific habit ID
				UserID:  currentUserID,
				Action:  habits.ActionHabitListView,
				Metadata: map[string]interface{}{
					"viewed_user_id": userID.String(),
					"count":          len(habitsData),
					"via":            "api",
					"path":           c.Request.URL.Path,
					"method":         c.Request.Method,
				},
			})
		}()
	}

	responses := make([]dto.HabitResponse, len(habitsData))
	for i, habit := range habitsData {
		responses[i] = *HabitToResponse(&habit)
	}

	c.JSON(http.StatusOK, gin.H{"data": responses})
}
