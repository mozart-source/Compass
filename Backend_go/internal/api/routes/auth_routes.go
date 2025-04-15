package routes

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// AuthRoutes handles the setup of auth-related routes
type AuthRoutes struct {
	handler   *handlers.AuthHandler
	jwtSecret string
}

// NewAuthRoutes creates a new AuthRoutes instance
func NewAuthRoutes(handler *handlers.AuthHandler, jwtSecret string) *AuthRoutes {
	return &AuthRoutes{
		handler:   handler,
		jwtSecret: jwtSecret,
	}
}

// RegisterRoutes registers all auth-related routes
func (ar *AuthRoutes) RegisterRoutes(router *gin.Engine) {
	// Create a roles group with authentication middleware
	rolesGroup := router.Group("/api/roles")
	rolesGroup.Use(middleware.NewAuthMiddleware(ar.jwtSecret))

	// Role management endpoints
	rolesGroup.POST("", middleware.RequirePermissions("roles:create"), ar.handler.CreateRole)
	rolesGroup.GET("", middleware.RequirePermissions("roles:read"), ar.handler.ListRoles)
	rolesGroup.GET("/:id", middleware.RequirePermissions("roles:read"), ar.handler.GetRole)
	rolesGroup.PUT("/:id", middleware.RequirePermissions("roles:update"), ar.handler.UpdateRole)
	rolesGroup.DELETE("/:id", middleware.RequirePermissions("roles:delete"), ar.handler.DeleteRole)

	// Role-Permission management endpoints
	rolesGroup.POST("/:id/permissions/:permission_id", middleware.RequirePermissions("roles:update"), ar.handler.AssignPermissionToRole)
	rolesGroup.DELETE("/:id/permissions/:permission_id", middleware.RequirePermissions("roles:update"), ar.handler.RemovePermissionFromRole)

	// User-Role management endpoints
	userRolesGroup := router.Group("/api/users")
	userRolesGroup.Use(middleware.NewAuthMiddleware(ar.jwtSecret))
	userRolesGroup.POST("/:user_id/roles/:role_id", middleware.RequirePermissions("roles:assign"), ar.handler.AssignRoleToUser)
	userRolesGroup.GET("/:user_id/roles", middleware.RequirePermissions("roles:read"), ar.handler.GetUserRoles)
}
