package handlers

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/user"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/security/auth"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/security/mfa"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MFAHandler handles MFA-related operations
type MFAHandler struct {
	userService user.Service
	jwtSecret   string
	logger      *logrus.Logger
}

// NewMFAHandler creates a new MFA handler
func NewMFAHandler(userService user.Service, jwtSecret string, logger *logrus.Logger) *MFAHandler {
	if logger == nil {
		logger = logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	return &MFAHandler{
		userService: userService,
		jwtSecret:   jwtSecret,
		logger:      logger,
	}
}

// SetupMFA sets up MFA for a user
// @Summary Setup MFA
// @Description Generate MFA secret and QR code for a user
// @Tags mfa
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MFASetupResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/mfa/setup [post]
func (h *MFAHandler) SetupMFA(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Setup MFA for the user
	setupResponse, err := h.userService.SetupMFA(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.WithError(err).Error("Failed to setup MFA")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup MFA"})
		return
	}

	// Ensure QR code is properly encoded as Base64
	qrCodeBase64 := setupResponse.QRCodeBase64
	if qrCodeBase64 != "" && !isBase64(qrCodeBase64) {
		qrCodeBase64 = base64.StdEncoding.EncodeToString([]byte(qrCodeBase64))
	}

	// Add data URI prefix if not present
	if qrCodeBase64 != "" && !hasDataUriPrefix(qrCodeBase64) {
		qrCodeBase64 = "data:image/png;base64," + qrCodeBase64
	}

	// Create response with properly formatted QR code
	response := dto.MFASetupResponse{
		Secret:       setupResponse.Secret,
		QRCodeBase64: qrCodeBase64,
		OTPAuthURL:   setupResponse.OTPAuthURL,
		BackupCodes:  setupResponse.BackupCodes,
	}

	c.JSON(http.StatusOK, response)
}

// isBase64 checks if a string is base64 encoded
func isBase64(s string) bool {
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil
}

// hasDataUriPrefix checks if a string has a data URI prefix
func hasDataUriPrefix(s string) bool {
	return len(s) > 5 && s[0:5] == "data:"
}

// VerifyMFA verifies and enables MFA for a user
// @Summary Verify and Enable MFA
// @Description Verify TOTP code and enable MFA for a user
// @Tags mfa
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code body dto.VerifyMFARequest true "TOTP verification code"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/mfa/verify [post]
func (h *MFAHandler) VerifyMFA(c *gin.Context) {
	var request dto.VerifyMFARequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify and enable MFA
	err := h.userService.VerifyAndEnableMFA(c.Request.Context(), userID.(uuid.UUID), request.Code)
	if err != nil {
		h.logger.WithError(err).Error("Failed to verify and enable MFA")

		if err == mfa.ErrInvalidCode {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid verification code"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable MFA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "MFA enabled successfully"})
}

// ValidateMFA validates an MFA code
// @Summary Validate MFA Code
// @Description Validate a TOTP code for an MFA-enabled user
// @Tags mfa
// @Accept json
// @Produce json
// @Param request body dto.ValidateMFARequest true "MFA validation request"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/mfa/validate [post]
func (h *MFAHandler) ValidateMFA(c *gin.Context) {
	var request dto.ValidateMFARequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse userID from string
	userID, err := uuid.Parse(request.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Validate the MFA code
	valid, err := h.userService.ValidateMFACode(c.Request.Context(), userID, request.Code)
	if err != nil {
		h.logger.WithError(err).Error("MFA validation error")
		if err == mfa.ErrInvalidCode {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid MFA code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to validate MFA code"})
		return
	}

	if !valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid MFA code"})
		return
	}

	// Get the user
	user, err := h.userService.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Error retrieving user after MFA validation")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete authentication"})
		return
	}

	// Get user's roles and permissions
	roles, permissions, err := h.userService.GetUserRolesAndPermissions(c.Request.Context(), user.ID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user roles and permissions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user permissions"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(
		user.ID,
		user.Email,
		roles,
		uuid.Nil,
		permissions,
		h.jwtSecret,
		24,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to generate token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Create session with device info
	session := auth.GetSessionStore().CreateSession(
		user.ID,
		c.Request.UserAgent(),
		c.ClientIP(),
		token,
		24*time.Hour,
	)

	response := dto.LoginResponse{
		Token:     token,
		ExpiresAt: session.ExpiresAt,
		User: dto.UserResponse{
			ID:          user.ID,
			Email:       user.Email,
			Username:    user.Username,
			FirstName:   user.FirstName,
			LastName:    user.LastName,
			PhoneNumber: user.PhoneNumber,
			AvatarURL:   user.AvatarURL,
			Bio:         user.Bio,
			Timezone:    user.Timezone,
			Locale:      user.Locale,
			IsActive:    user.IsActive,
			IsSuperuser: user.IsSuperuser,
			MFAEnabled:  user.MFAEnabled,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
		Session: dto.SessionResponse{
			ID:           session.ID,
			DeviceInfo:   session.DeviceInfo,
			IPAddress:    session.IPAddress,
			LastActivity: session.LastActivity,
			ExpiresAt:    session.ExpiresAt,
		},
	}

	c.JSON(http.StatusOK, response)
}

// DisableMFA disables MFA for a user
// @Summary Disable MFA
// @Description Disable MFA for a user
// @Tags mfa
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param password body dto.DisableMFARequest true "User's current password"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/mfa/disable [post]
func (h *MFAHandler) DisableMFA(c *gin.Context) {
	var request dto.DisableMFARequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Disable MFA
	err := h.userService.DisableMFA(c.Request.Context(), userID.(uuid.UUID), request.Password)
	if err != nil {
		h.logger.WithError(err).Error("Failed to disable MFA")

		if err == user.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable MFA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "MFA disabled successfully"})
}

// GetMFAStatus retrieves the MFA status for a user
// @Summary Get MFA Status
// @Description Get current MFA status for a user
// @Tags mfa
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.MFAStatusResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/mfa/status [get]
func (h *MFAHandler) GetMFAStatus(c *gin.Context) {
	// Get user ID from JWT token
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get MFA status
	enabled, err := h.userService.IsMFAEnabled(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		h.logger.WithError(err).Error("Failed to get MFA status")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get MFA status"})
		return
	}

	c.JSON(http.StatusOK, dto.MFAStatusResponse{Enabled: enabled})
}
