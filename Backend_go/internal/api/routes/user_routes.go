package routes

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/security/auth"
	"github.com/gin-gonic/gin"
)

type UserRoutes struct {
	userHandler *handlers.UserHandler
	jwtSecret   string
	rateLimiter *auth.RedisRateLimiter
}

func NewUserRoutes(userHandler *handlers.UserHandler, jwtSecret string, rateLimiter *auth.RedisRateLimiter) *UserRoutes {
	return &UserRoutes{
		userHandler: userHandler,
		jwtSecret:   jwtSecret,
		rateLimiter: rateLimiter,
	}
}

// RegisterRoutes sets up all user-related routes
func (ur *UserRoutes) RegisterRoutes(router *gin.Engine) {
	// Create validation middleware instance
	validation := middleware.NewValidationMiddleware()

	userGroup := router.Group("/api/users")
	{
		// Public routes with stricter rate limiting
		public := userGroup.Group("")
		public.Use(middleware.RateLimitMiddleware(ur.rateLimiter))
		{
			// Apply validation to registration and login
			public.POST("/register", validation.ValidateRequest(&dto.CreateUserRequest{}), ur.userHandler.CreateUser)
			public.POST("/login", validation.ValidateRequest(&dto.LoginRequest{}), ur.userHandler.Login)
		}

		// Protected routes with general API rate limiting
		protected := userGroup.Group("")
		protected.Use(
			middleware.NewAuthMiddleware(ur.jwtSecret),
			middleware.RateLimitMiddleware(ur.rateLimiter),
		)
		{
			// Profile management
			protected.GET("/profile", ur.userHandler.GetUser)
			protected.PUT("/profile", validation.ValidateRequest(&dto.UpdateUserRequest{}), ur.userHandler.UpdateUser)
			protected.DELETE("/profile", ur.userHandler.DeleteUser)

			// Session management
			protected.GET("/sessions", ur.userHandler.GetUserSessions)
			protected.POST("/sessions/:id/revoke", ur.userHandler.RevokeSession)
			protected.POST("/logout", ur.userHandler.Logout)

			// Analytics routes
			analyticsGroup := protected.Group("/analytics")
			{
				// User activity
				analyticsGroup.GET("/activity", ur.userHandler.GetUserActivity)
				analyticsGroup.POST("/record", validation.ValidateRequest(&dto.CreateUserRequest{}), ur.userHandler.RecordUserActivity)

				// Session activity
				analyticsGroup.GET("/sessions", ur.userHandler.GetSessionActivity)

				// Summary
				analyticsGroup.GET("/summary", ur.userHandler.GetUserActivitySummary)
			}
		}
	}
}
