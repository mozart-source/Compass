package routes

import (
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// TaskRoutes handles the setup of task-related routes
type TaskRoutes struct {
	handler   *handlers.TaskHandler
	jwtSecret string
}

// NewTaskRoutes creates a new TaskRoutes instance
func NewTaskRoutes(handler *handlers.TaskHandler, jwtSecret string) *TaskRoutes {
	return &TaskRoutes{
		handler:   handler,
		jwtSecret: jwtSecret,
	}
}

// RegisterRoutes registers all task-related routes
func (r *TaskRoutes) RegisterRoutes(router *gin.Engine, cache *middleware.CacheMiddleware) {
	// Initialize task-specific middleware
	validation := middleware.NewValidationMiddleware()
	metrics := middleware.NewMetricsMiddleware()
	circuitBreaker := middleware.NewCircuitBreaker(middleware.CircuitBreakerConfig{
		FailureThreshold:    5,
		SuccessThreshold:    2,
		Timeout:             15 * time.Second,
		HalfOpenMaxRequests: 3,
	})

	tasks := router.Group("/api/tasks")
	tasks.Use(middleware.NewAuthMiddleware(r.jwtSecret))
	tasks.Use(metrics.CollectMetrics())

	// Apply circuit breaker to task operations to prevent cascading failures
	tasks.Use(circuitBreaker.CircuitBreakerMiddleware())

	// Read operations with caching
	tasks.GET("", cache.CacheResponse(), r.handler.ListTasks)
	tasks.GET("/:id", cache.CacheResponse(), r.handler.GetTask)
	tasks.GET("/user/:user_id", cache.CacheResponse(), r.handler.ListTasks)
	tasks.GET("/project/:project_id", cache.CacheResponse(), r.handler.GetProjectTasks)

	// Write operations with cache invalidation and validation
	tasks.POST("", validation.ValidateRequest(&dto.CreateTaskRequest{}), cache.CacheInvalidate("tasks:*"), r.handler.CreateTask)
	tasks.PUT("/:id", validation.ValidateRequest(&dto.UpdateTaskRequest{}), cache.CacheInvalidate("tasks:*"), r.handler.UpdateTask)
	tasks.DELETE("/:id", cache.CacheInvalidate("tasks:*"), r.handler.DeleteTask)

	// Status updates
	tasks.PATCH("/:id/status", validation.ValidateRequest(&dto.UpdateTaskStatusRequest{}), cache.CacheInvalidate("tasks:*"), r.handler.UpdateTaskStatus)
	tasks.PATCH("/:id/assign", validation.ValidateRequest(&dto.AssignTaskRequest{}), cache.CacheInvalidate("tasks:*"), r.handler.AssignTask)

	// Task analytics routes
	analytics := tasks.Group("/analytics")

	// User-specific analytics
	analytics.GET("/user", r.handler.GetUserTaskAnalytics)
	analytics.GET("/user/summary", r.handler.GetUserTaskActivitySummary)

	// Task-specific analytics
	tasks.GET("/:id/analytics", r.handler.GetTaskAnalytics)
	tasks.GET("/:id/analytics/summary", r.handler.GetTaskActivitySummary)
	tasks.POST("/:id/analytics/record", validation.ValidateRequest(&dto.RecordUserActivityRequest{}), r.handler.RecordTaskActivity)
}
