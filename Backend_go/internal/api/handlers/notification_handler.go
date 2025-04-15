package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/notification"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/logger"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/security/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// NotificationHandler handles notification-related requests
type NotificationHandler struct {
	service  notification.Service
	logger   *logger.Logger
	upgrader websocket.Upgrader
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(service notification.Service, logger *logger.Logger) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		logger:  logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}
}

// GetAll godoc
// @Summary Get all notifications for a user
// @Description Get all notifications for the authenticated user with pagination
// @Tags notifications
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Security BearerAuth
// @Success 200 {object} dto.NotificationListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/notifications [get]
func (h *NotificationHandler) GetAll(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, err := uuid.Parse(userID.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if pageVal, err := strconv.Atoi(pageStr); err == nil && pageVal > 0 {
			page = pageVal
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSizeVal, err := strconv.Atoi(pageSizeStr); err == nil && pageSizeVal > 0 && pageSizeVal <= 100 {
			pageSize = pageSizeVal
		}
	}

	offset := (page - 1) * pageSize

	notifications, err := h.service.GetByUserID(c.Request.Context(), uid, pageSize, offset)
	if err != nil {
		h.logger.Error("Failed to get notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	unreadCount, err := h.service.CountUnread(c.Request.Context(), uid)
	if err != nil {
		h.logger.Error("Failed to count unread notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count unread notifications"})
		return
	}

	response := dto.NotificationListResponse{
		Items:       dto.ToDTOs(notifications),
		TotalCount:  len(notifications), // This should ideally be from a separate count query
		UnreadCount: unreadCount,
		Page:        page,
		PageSize:    pageSize,
	}

	c.JSON(http.StatusOK, response)
}

// GetUnread godoc
// @Summary Get unread notifications
// @Description Get unread notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Security BearerAuth
// @Success 200 {object} dto.NotificationListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/notifications/unread [get]
func (h *NotificationHandler) GetUnread(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, err := uuid.Parse(userID.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if pageVal, err := strconv.Atoi(pageStr); err == nil && pageVal > 0 {
			page = pageVal
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSizeVal, err := strconv.Atoi(pageSizeStr); err == nil && pageSizeVal > 0 && pageSizeVal <= 100 {
			pageSize = pageSizeVal
		}
	}

	offset := (page - 1) * pageSize

	notifications, err := h.service.GetUnreadByUserID(c.Request.Context(), uid, pageSize, offset)
	if err != nil {
		h.logger.Error("Failed to get unread notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread notifications"})
		return
	}

	unreadCount, err := h.service.CountUnread(c.Request.Context(), uid)
	if err != nil {
		h.logger.Error("Failed to count unread notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count unread notifications"})
		return
	}

	response := dto.NotificationListResponse{
		Items:       dto.ToDTOs(notifications),
		TotalCount:  unreadCount,
		UnreadCount: unreadCount,
		Page:        page,
		PageSize:    pageSize,
	}

	c.JSON(http.StatusOK, response)
}

// GetByID godoc
// @Summary Get notification by ID
// @Description Get a notification by its ID
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Security BearerAuth
// @Success 200 {object} dto.NotificationDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/notifications/{id} [get]
func (h *NotificationHandler) GetByID(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	notificationID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	notif, err := h.service.GetByID(c.Request.Context(), notificationID)
	if err != nil {
		var notFoundErr *notification.ErrNotFoundType
		if errors.As(err, &notFoundErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		h.logger.Error("Failed to get notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification"})
		return
	}

	// Check if notification belongs to the authenticated user
	if notif.UserID.String() != userID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this notification"})
		return
	}

	c.JSON(http.StatusOK, dto.ToDTO(notif))
}

// MarkAsRead godoc
// @Summary Mark notification as read
// @Description Mark a notification as read
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	// Get validated model from context if available
	var updateReq dto.NotificationUpdateRequest
	validatedModel, exists := c.Get("validated_model")

	if exists {
		// If validation middleware provided the model, use it
		if validatedPtr, ok := validatedModel.(*dto.NotificationUpdateRequest); ok {
			updateReq = *validatedPtr
		} else {
			h.logger.Error("Invalid model type from validation", zap.Any("type", validatedModel))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}
	} else if c.Request.ContentLength > 0 {
		// Only try to bind if there's actually content
		if err := c.ShouldBindJSON(&updateReq); err != nil {
			h.logger.Error("Failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	notificationID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	// Get notification first to check permission
	notif, err := h.service.GetByID(c.Request.Context(), notificationID)
	if err != nil {
		var notFoundErr *notification.ErrNotFoundType
		if errors.As(err, &notFoundErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		h.logger.Error("Failed to get notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification"})
		return
	}

	// Check if notification belongs to the authenticated user
	if notif.UserID.String() != userID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to modify this notification"})
		return
	}

	if err := h.service.MarkAsRead(c.Request.Context(), notificationID); err != nil {
		h.logger.Error("Failed to mark notification as read", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

// MarkAllAsRead godoc
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, err := uuid.Parse(userID.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.service.MarkAllAsRead(c.Request.Context(), uid); err != nil {
		h.logger.Error("Failed to mark all notifications as read", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// CountUnread godoc
// @Summary Count unread notifications
// @Description Count unread notifications for the authenticated user
// @Tags notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.NotificationCountResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/notifications/count [get]
func (h *NotificationHandler) CountUnread(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	uid, err := uuid.Parse(userID.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	unreadCount, err := h.service.CountUnread(c.Request.Context(), uid)
	if err != nil {
		h.logger.Error("Failed to count unread notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count unread notifications"})
		return
	}

	c.JSON(http.StatusOK, dto.NotificationCountResponse{
		UnreadCount: unreadCount,
		TotalCount:  0, // This should ideally be from a separate count query
	})
}

// Delete godoc
// @Summary Delete notification
// @Description Delete a notification
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path string true "Notification ID"
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/notifications/{id} [delete]
func (h *NotificationHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	id := c.Param("id")
	notificationID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	// Get notification first to check permission
	notif, err := h.service.GetByID(c.Request.Context(), notificationID)
	if err != nil {
		var notFoundErr *notification.ErrNotFoundType
		if errors.As(err, &notFoundErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		h.logger.Error("Failed to get notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification"})
		return
	}

	// Check if notification belongs to the authenticated user
	if notif.UserID.String() != userID.String() {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete this notification"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), notificationID); err != nil {
		h.logger.Error("Failed to delete notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted"})
}

// Create godoc
// @Summary Create notification
// @Description Create a new notification (admin only)
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body dto.CreateNotificationRequest true "Notification data"
// @Security BearerAuth
// @Success 201 {object} dto.NotificationDTO
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/notifications [post]
func (h *NotificationHandler) Create(c *gin.Context) {
	var req dto.CreateNotificationRequest
	validatedModel, exists := c.Get("validated_model")

	if exists {
		// If validation middleware provided the model, use it
		if validatedPtr, ok := validatedModel.(*dto.CreateNotificationRequest); ok {
			req = *validatedPtr
		} else {
			h.logger.Error("Invalid model type from validation", zap.Any("type", validatedModel))
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}
	} else {
		// If validation middleware didn't run, do manual binding
		if err := c.ShouldBindJSON(&req); err != nil {
			h.logger.Error("Failed to bind request", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Check if this is admin or has proper permission (optional)
	// isAdmin := c.GetBool("isAdmin")
	// if !isAdmin {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can create notifications"})
	// 	return
	// }

	notif := req.ToModel()
	if err := h.service.Create(c.Request.Context(), notif); err != nil {
		h.logger.Error("Failed to create notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	c.JSON(http.StatusCreated, dto.ToDTO(notif))
}

// WebSocketHandler handles WebSocket connections for real-time notifications
func (h *NotificationHandler) WebSocketHandler(c *gin.Context) {
	// Check if we already have user_id from auth middleware
	userID, exists := middleware.GetUserID(c)

	// If not authenticated via normal middleware, try to get from query parameter
	if !exists {
		// Try to get token from query parameter
		tokenParam := c.Query("token")
		if tokenParam == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		// Validate token from query parameter
		jwtSecret := c.MustGet("jwt_secret").(string)
		claims, err := auth.ValidateToken(tokenParam, jwtSecret)
		if err != nil {
			h.logger.Error("WebSocket token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// Use the user ID from the validated token
		userID = claims.UserID
	}

	uid, err := uuid.Parse(userID.String())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Log the WebSocket connection attempt
	h.logger.Info("WebSocket connection attempt",
		zap.String("user_id", uid.String()),
		zap.String("remote_addr", c.Request.RemoteAddr))

	// Upgrade the connection with enhanced error handling
	ws, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade to WebSocket",
			zap.Error(err),
			zap.String("user_id", uid.String()),
			zap.String("remote_addr", c.Request.RemoteAddr))
		return
	}
	defer func() {
		ws.Close()
		h.logger.Info("WebSocket connection closed", zap.String("user_id", uid.String()))
	}()

	// Configure WebSocket
	ws.SetReadLimit(1024 * 10) // 10KB message size limit
	ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.SetPongHandler(func(string) error {
		ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Subscribe to notifications for this user
	notifChan, cancel, err := h.service.SubscribeToNotifications(uid)
	if err != nil {
		h.logger.Error("Failed to subscribe to notifications",
			zap.Error(err),
			zap.String("user_id", uid.String()))
		ws.WriteJSON(map[string]interface{}{
			"error": "Failed to subscribe to notifications",
		})
		return
	}
	defer cancel()

	// Send initial count of unread notifications
	unreadCount, err := h.service.CountUnread(c.Request.Context(), uid)
	if err != nil {
		h.logger.Error("Failed to count unread notifications",
			zap.Error(err),
			zap.String("user_id", uid.String()))
	} else {
		countMsg := map[string]interface{}{
			"type":  "count",
			"count": unreadCount,
		}
		if writeErr := ws.WriteJSON(countMsg); writeErr != nil {
			h.logger.Error("Failed to send initial count", zap.Error(writeErr))
			return
		}
	}

	// Create a ping ticker to keep connection alive
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// Create a channel for WebSocket control messages
	done := make(chan struct{})

	// Handle incoming messages (like read receipts)
	go func() {
		defer close(done)
		for {
			// Read WebSocket message
			messageType, message, err := ws.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err,
					websocket.CloseGoingAway,
					websocket.CloseAbnormalClosure) {
					h.logger.Error("WebSocket read error",
						zap.Error(err),
						zap.String("user_id", uid.String()))
				}
				return
			}

			// Handle ping/pong
			if messageType == websocket.PingMessage {
				if err := ws.WriteMessage(websocket.PongMessage, nil); err != nil {
					h.logger.Error("WebSocket pong write error", zap.Error(err))
					return
				}
				continue
			}

			// Handle other message types
			if messageType == websocket.TextMessage && len(message) > 0 {
				// Try to parse as JSON
				var msgData map[string]interface{}
				if jsonErr := json.Unmarshal(message, &msgData); jsonErr == nil {
					// Handle command messages
					if cmd, ok := msgData["command"].(string); ok {
						switch cmd {
						case "mark_read":
							if id, ok := msgData["id"].(string); ok {
								notifID, parseErr := uuid.Parse(id)
								if parseErr == nil {
									h.service.MarkAsRead(c.Request.Context(), notifID)
								}
							}
						case "mark_all_read":
							h.service.MarkAllAsRead(c.Request.Context(), uid)
						}
					}
				}
			}

			h.logger.Info("Received WebSocket message",
				zap.Binary("message", message),
				zap.String("user_id", uid.String()))
		}
	}()

	// Main loop for sending notifications
	for {
		select {
		case notification, ok := <-notifChan:
			// Channel closed
			if !ok {
				return
			}

			// Send notification to WebSocket
			if err := ws.WriteJSON(dto.ToDTO(notification)); err != nil {
				h.logger.Error("WebSocket write error",
					zap.Error(err),
					zap.String("user_id", uid.String()))
				return
			}

		case <-pingTicker.C:
			// Send ping to keep connection alive
			if err := ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				h.logger.Error("WebSocket ping error",
					zap.Error(err),
					zap.String("user_id", uid.String()))
				return
			}

		case <-done:
			// WebSocket closed by client
			return
		}
	}
}
