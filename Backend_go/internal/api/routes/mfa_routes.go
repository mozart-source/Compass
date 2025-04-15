package routes

import (
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/handlers"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// MFARoutes defines routes for MFA operations
type MFARoutes struct {
	mfaHandler *handlers.MFAHandler
	jwtSecret  string
}

// NewMFARoutes creates a new MFA routes instance
func NewMFARoutes(mfaHandler *handlers.MFAHandler, jwtSecret string) *MFARoutes {
	return &MFARoutes{
		mfaHandler: mfaHandler,
		jwtSecret:  jwtSecret,
	}
}

// RegisterRoutes registers MFA routes with the given router
func (r *MFARoutes) RegisterRoutes(router *gin.Engine) {
	// Create validation middleware instance
	validation := middleware.NewValidationMiddleware()

	// MFA routes (all protected)
	mfaGroup := router.Group("/api/users/mfa")
	mfaGroup.Use(middleware.NewAuthMiddleware(r.jwtSecret))
	{
		// Setup MFA (generates QR code)
		mfaGroup.POST("/setup", r.mfaHandler.SetupMFA)

		// Verify and enable MFA
		mfaGroup.POST("/verify", validation.ValidateRequest(&dto.VerifyMFARequest{}), r.mfaHandler.VerifyMFA)

		// Disable MFA
		mfaGroup.POST("/disable", validation.ValidateRequest(&dto.DisableMFARequest{}), r.mfaHandler.DisableMFA)

		// Get MFA status
		mfaGroup.GET("/status", r.mfaHandler.GetMFAStatus)
	}

	// Public MFA validation endpoint for login
	authGroup := router.Group("/api/auth/mfa")
	{
		// Validate MFA code during login
		authGroup.POST("/validate", validation.ValidateRequest(&dto.ValidateMFARequest{}), r.mfaHandler.ValidateMFA)
	}
}
