package routes

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

type HabitNotificationRoutes struct {
	handler   *handlers.HabitNotificationHandler
	jwtSecret string
}

func NewHabitNotificationRoutes(handler *handlers.HabitNotificationHandler, jwtSecret string) *HabitNotificationRoutes {
	return &HabitNotificationRoutes{
		handler:   handler,
		jwtSecret: jwtSecret,
	}
}

// RegisterRoutes registers all habit notification-related routes
func (h *HabitNotificationRoutes) RegisterRoutes(router *gin.Engine, cache *middleware.CacheMiddleware) {
	// Register under habits route for specific habit-related notifications
	habitNotifications := router.Group("/api/habits")
	habitNotifications.Use(middleware.NewAuthMiddleware(h.jwtSecret))

	// Get all notifications for a specific habit
	habitNotifications.GET("/:id/notifications", cache.CacheResponse(), h.handler.GetNotificationsByHabit)

	// Create a custom notification for a habit
	habitNotifications.POST("/:id/notifications", cache.CacheInvalidate("notifications:*"), h.handler.CreateCustomHabitNotification)
}
