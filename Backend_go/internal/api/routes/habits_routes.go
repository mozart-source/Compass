package routes

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

type HabitsRoutes struct {
	handler   *handlers.HabitsHandler
	jwtSecret string
}

func NewHabitsRoutes(handler *handlers.HabitsHandler, jwtSecret string) *HabitsRoutes {
	return &HabitsRoutes{
		handler:   handler,
		jwtSecret: jwtSecret,
	}
}

// RegisterRoutes registers all habit-related routes
// @Summary Register habits routes
// @Description Register all habit-related routes with their handlers
// @Tags habits
// @Security BearerAuth
func (h *HabitsRoutes) RegisterRoutes(router *gin.Engine, cache *middleware.CacheMiddleware) {
	// Initialize middleware components
	validation := middleware.NewValidationMiddleware()
	circuitBreaker := middleware.NewCircuitBreaker(middleware.CircuitBreakerConfig{
		FailureThreshold:    3,
		SuccessThreshold:    2,
		Timeout:             30,
		HalfOpenMaxRequests: 5,
	})

	habits := router.Group("/api/habits")
	habits.Use(middleware.NewAuthMiddleware(h.jwtSecret))

	// Apply circuit breaker to the entire habits group - critical for maintaining system stability
	habits.Use(circuitBreaker.CircuitBreakerMiddleware())

	// List and filter - specific routes first
	// Apply compression for large data responses like heatmaps and stats
	habits.GET("", cache.CacheResponse(), gzip.Gzip(gzip.DefaultCompression), h.handler.ListHabits)
	habits.POST("", validation.ValidateRequest(&dto.CreateHabitRequest{}), cache.CacheInvalidate("habits:*"), h.handler.CreateHabit)
	habits.GET("/heatmap", cache.CacheResponse(), gzip.Gzip(gzip.DefaultCompression), h.handler.GetHabitHeatmap)
	habits.GET("/due-today", cache.CacheResponse(), gzip.Gzip(gzip.DefaultCompression), h.handler.GetHabitsDueToday)
	habits.GET("/user/:user_id", cache.CacheResponse(), gzip.Gzip(gzip.DefaultCompression), h.handler.GetUserHabits)

	// Analytics routes
	analytics := habits.Group("/analytics")
	analytics.GET("/user", h.handler.GetUserHabitAnalytics)
	analytics.GET("/user/summary", h.handler.GetUserHabitActivitySummary)

	// CRUD operations with parameters
	habits.GET("/:id", cache.CacheResponse(), gzip.Gzip(gzip.DefaultCompression), h.handler.GetHabit)
	habits.PUT("/:id", validation.ValidateRequest(&dto.UpdateHabitRequest{}), cache.CacheInvalidate("habits:*"), h.handler.UpdateHabit)
	habits.DELETE("/:id", cache.CacheInvalidate("habits:*"), h.handler.DeleteHabit)

	// Habit completion routes
	habits.POST("/:id/complete", cache.CacheInvalidate("habits:*"), h.handler.MarkHabitCompleted)
	habits.POST("/:id/uncomplete", cache.CacheInvalidate("habits:*"), h.handler.UnmarkHabitCompleted)
	habits.GET("/:id/stats", cache.CacheResponse(), h.handler.GetHabitStats)
	habits.GET("/:id/streak-history", cache.CacheResponse(), h.handler.GetStreakHistory)

	// Per-habit analytics routes
	habits.GET("/:id/analytics", h.handler.GetHabitAnalytics)
	habits.GET("/:id/analytics/summary", h.handler.GetHabitActivitySummary)
	habits.POST("/:id/analytics/record", validation.ValidateRequest(&dto.RecordHabitActivityRequest{}), h.handler.RecordHabitActivity)
}
