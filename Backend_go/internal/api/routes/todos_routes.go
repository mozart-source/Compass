package routes

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// TodosRoutes handles the setup of todo-related routes
type TodosRoutes struct {
	handler   *handlers.TodoHandler
	jwtSecret string
}

// NewTodosRoutes creates a new TodosRoutes instance
func NewTodosRoutes(handler *handlers.TodoHandler, jwtSecret string) *TodosRoutes {
	return &TodosRoutes{
		handler:   handler,
		jwtSecret: jwtSecret,
	}
}

// RegisterRoutes registers all todo-related routes
func (r *TodosRoutes) RegisterRoutes(router *gin.Engine, cache *middleware.CacheMiddleware) {
	todos := router.Group("/api/todos")
	todos.Use(middleware.NewAuthMiddleware(r.jwtSecret))

	// Read operations with caching
	todos.GET("", cache.CacheResponse(), r.handler.ListTodos)
	todos.GET("/:id", cache.CacheResponse(), r.handler.GetTodo)
	todos.GET("/user/:user_id", cache.CacheResponse(), r.handler.GetTodosByUser)

	// Write operations with cache invalidation - invalidate both todos and todo-lists
	todos.POST("", cache.CacheInvalidate("todos:*", "todo-lists:*"), r.handler.CreateTodo)
	todos.PUT("/:id", cache.CacheInvalidate("todos:*", "todo-lists:*"), r.handler.UpdateTodo)
	todos.DELETE("/:id", cache.CacheInvalidate("todos:*", "todo-lists:*"), r.handler.DeleteTodo)

	// Status and priority updates - invalidate both todos and todo-lists
	todos.PATCH("/:id/status", cache.CacheInvalidate("todos:*", "todo-lists:*"), r.handler.UpdateTodoStatus)
	todos.PATCH("/:id/priority", cache.CacheInvalidate("todos:*", "todo-lists:*"), r.handler.UpdateTodoPriority)

	// Completion status - invalidate both todos and todo-lists
	todos.PATCH("/:id/complete", cache.CacheInvalidate("todos:*", "todo-lists:*"), r.handler.CompleteTodo)
	todos.PATCH("/:id/uncomplete", cache.CacheInvalidate("todos:*", "todo-lists:*"), r.handler.UncompleteTodo)

	// Todo Lists routes
	todoLists := router.Group("/api/todo-lists")
	todoLists.Use(middleware.NewAuthMiddleware(r.jwtSecret))

	// Read operations with caching
	todoLists.GET("", cache.CacheResponse(), r.handler.GetAllTodoLists)
	todoLists.GET("/:id", cache.CacheResponse(), r.handler.GetTodoList)

	// Write operations with cache invalidation
	todoLists.POST("", cache.CacheInvalidate("todos:*", "todo-lists:*"), r.handler.CreateTodoList)
	todoLists.PUT("/:id", cache.CacheInvalidate("todos:*", "todo-lists:*"), r.handler.UpdateTodoList)
	todoLists.DELETE("/:id", cache.CacheInvalidate("todos:*", "todo-lists:*"), r.handler.DeleteTodoList)
}
