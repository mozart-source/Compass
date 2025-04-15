package handlers

import (
	"errors"
	"net/http"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/habits"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/notification"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// HabitNotificationHandler handles notification routes related to habits
type HabitNotificationHandler struct {
	habitService         habits.Service
	notificationService  notification.Service
	habitNotificationSvc *habits.HabitNotificationService
}

// NewHabitNotificationHandler creates a new habit notification handler
func NewHabitNotificationHandler(habitService habits.Service, notificationService notification.Service, habitNotificationSvc *habits.HabitNotificationService) *HabitNotificationHandler {
	return &HabitNotificationHandler{
		habitService:         habitService,
		notificationService:  notificationService,
		habitNotificationSvc: habitNotificationSvc,
	}
}

// GetNotificationsByHabit retrieves all notifications related to a specific habit
func (h *HabitNotificationHandler) GetNotificationsByHabit(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	habitIDStr := c.Param("id")
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid habit ID"})
		return
	}

	// Verify that user owns the habit
	habit, err := h.habitService.GetHabit(c, habitID)
	if err != nil {
		if errors.Is(err, habits.ErrHabitNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve habit"})
		}
		return
	}

	if habit.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to view this habit's notifications"})
		return
	}

	// Parse filter parameters
	var filter dto.NotificationFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter parameters"})
		return
	}

	// Get all notifications for the user
	limit := filter.PageSize
	offset := filter.Page * filter.PageSize
	notifications, err := h.notificationService.GetByUserID(c, userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve notifications"})
		return
	}

	// Filter notifications for this specific habit
	var habitNotifications []*notification.Notification
	for _, n := range notifications {
		if n.Reference == "habits" && n.ReferenceID == habitID {
			habitNotifications = append(habitNotifications, n)
		}
	}

	// Get unread count
	unreadCount, err := h.notificationService.CountUnread(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count unread notifications"})
		return
	}

	// Convert domain notifications to DTOs
	notificationDTOs := dto.ToDTOs(habitNotifications)

	c.JSON(http.StatusOK, gin.H{
		"items":        notificationDTOs,
		"total_count":  len(habitNotifications),
		"unread_count": unreadCount,
		"page":         filter.Page,
		"page_size":    filter.PageSize,
	})
}

// CreateCustomHabitNotification creates a custom notification for a habit
func (h *HabitNotificationHandler) CreateCustomHabitNotification(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	habitIDStr := c.Param("id")
	habitID, err := uuid.Parse(habitIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid habit ID"})
		return
	}

	// Verify that user owns the habit
	habit, err := h.habitService.GetHabit(c, habitID)
	if err != nil {
		if errors.Is(err, habits.ErrHabitNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Habit not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve habit"})
		}
		return
	}

	if habit.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to create notifications for this habit"})
		return
	}

	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
		Type    string `json:"type" binding:"required,oneof=habit_completed habit_streak habit_broken habit_reminder habit_milestone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Create notification
	err = h.notificationService.CreateForUser(
		c,
		userID,
		notification.Type(req.Type),
		req.Title,
		req.Content,
		map[string]string{
			"habitID": habit.ID.String(),
			"title":   habit.Title,
		},
		"habits",
		habit.ID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Notification created successfully"})
}
