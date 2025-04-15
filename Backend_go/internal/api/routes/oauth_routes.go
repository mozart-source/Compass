package routes

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/security/auth"
	"github.com/gin-gonic/gin"
)

// OAuthRoutes defines the routes for OAuth2 authentication
type OAuthRoutes struct {
	handler     *handlers.OAuthHandler
	rateLimiter *auth.RedisRateLimiter
}

// NewOAuthRoutes creates a new OAuthRoutes instance
func NewOAuthRoutes(handler *handlers.OAuthHandler, rateLimiter *auth.RedisRateLimiter) *OAuthRoutes {
	return &OAuthRoutes{
		handler:     handler,
		rateLimiter: rateLimiter,
	}
}

// RegisterRoutes registers the OAuth2 routes
func (r *OAuthRoutes) RegisterRoutes(router *gin.Engine) {
	routes := router.Group("/api/auth/oauth")

	// Apply rate limiting middleware
	routes.Use(middleware.RateLimitMiddleware(r.rateLimiter))

	// Public routes - no auth required
	routes.GET("/providers", r.handler.GetProviders)

	// Support both GET and POST for login to be more flexible
	routes.POST("/login", r.handler.InitiateLogin)

	// Support both GET and POST for callback to be more flexible
	routes.POST("/callback", r.handler.HandleCallback)
	routes.GET("/callback", r.handler.HandleCallback) // for OAuth redirects
}
