package routes

import (
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/security/auth"
	"github.com/gin-gonic/gin"
)

// NotificationRoutes manages notification endpoint routes
type NotificationRoutes struct {
	handler     *handlers.NotificationHandler
	jwtSecret   string
	rateLimiter auth.RateLimiter
}

// NewNotificationRoutes creates a new notification routes handler
func NewNotificationRoutes(handler *handlers.NotificationHandler, jwtSecret string, rateLimiter auth.RateLimiter) *NotificationRoutes {
	return &NotificationRoutes{
		handler:     handler,
		jwtSecret:   jwtSecret,
		rateLimiter: rateLimiter,
	}
}

// RegisterRoutes registers notification routes with the provided router
func (r *NotificationRoutes) RegisterRoutes(router *gin.Engine, cacheMiddleware *middleware.CacheMiddleware) {
	// Initialize middleware components that are well-suited for notifications
	validation := middleware.NewValidationMiddleware()
	tracing := middleware.NewTracingMiddleware()

	// Create a route group with authentication middleware
	authMiddleware := middleware.NewAuthMiddleware(r.jwtSecret)

	// Apply moderate rate limiting to notification endpoints (more requests allowed than auth endpoints)
	notificationRateLimiter := middleware.RateLimitMiddleware(r.rateLimiter.WithLimit(120, time.Minute))

	// Set JWT secret middleware
	jwtSecretMiddleware := func(c *gin.Context) {
		c.Set("jwt_secret", r.jwtSecret)
		c.Next()
	}

	// Notification routes
	notificationRoutes := router.Group("/api/notifications")
	notificationRoutes.Use(authMiddleware)
	notificationRoutes.Use(notificationRateLimiter)
	notificationRoutes.Use(jwtSecretMiddleware)
	{
		// GET endpoints
		// Apply compression for endpoints that might return large datasets
		notificationRoutes.GET("", cacheMiddleware.CachePageWithTTL("notifications", 30*time.Second), r.handler.GetAll)
		notificationRoutes.GET("/unread", r.handler.GetUnread) // No cache for unread - always fresh
		notificationRoutes.GET("/count", r.handler.CountUnread)
		notificationRoutes.GET("/:id", cacheMiddleware.CachePageWithTTL("notification", 1*time.Minute), r.handler.GetByID)

		// PUT endpoints with validation
		notificationRoutes.PUT("/:id/read", validation.ValidateRequest(&dto.NotificationUpdateRequest{}), r.handler.MarkAsRead)
		notificationRoutes.PUT("/read-all", r.handler.MarkAllAsRead)

		// DELETE endpoint
		notificationRoutes.DELETE("/:id", r.handler.Delete)

		// POST endpoint (typically for admin or system use)
		notificationRoutes.POST("", validation.ValidateRequest(&dto.CreateNotificationRequest{}), r.handler.Create)
	}

	// WebSocket endpoint (no auth middleware, handles token via query parameter)
	// This needs to be registered separately to avoid the auth middleware
	wsRoute := router.Group("/api/notifications")
	wsRoute.Use(jwtSecretMiddleware)
	wsRoute.Use(tracing.TraceRequest())
	wsRoute.GET("/ws", r.handler.WebSocketHandler)
}
