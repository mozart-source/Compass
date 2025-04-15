package routes

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// OrganizationRoutes handles the setup of organization-related routes
type OrganizationRoutes struct {
	handler   *handlers.OrganizationHandler
	jwtSecret string
}

// NewOrganizationRoutes creates a new OrganizationRoutes instance
func NewOrganizationRoutes(handler *handlers.OrganizationHandler, jwtSecret string) *OrganizationRoutes {
	return &OrganizationRoutes{
		handler:   handler,
		jwtSecret: jwtSecret,
	}
}

// RegisterRoutes registers all organization-related routes
func (or *OrganizationRoutes) RegisterRoutes(router *gin.Engine) {
	// Create an organization group with authentication middleware
	organizationGroup := router.Group("/api/organizations")
	organizationGroup.Use(middleware.NewAuthMiddleware(or.jwtSecret))

	organizationGroup.POST("", or.handler.CreateOrganization)
	organizationGroup.GET("", or.handler.ListOrganizations)
	organizationGroup.GET("/:id", or.handler.GetOrganization)
	organizationGroup.GET("/:id/stats", or.handler.GetOrganizationStats)
	organizationGroup.PUT("/:id", or.handler.UpdateOrganization)
	organizationGroup.DELETE("/:id", or.handler.DeleteOrganization)
}
