package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/user"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/security/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OAuthHandler handles OAuth2 authentication
type OAuthHandler struct {
	oauthService *auth.OAuthService
	userService  user.Service
	jwtSecret    string
	logger       *zap.Logger
}

// NewOAuthHandler creates a new OAuthHandler
func NewOAuthHandler(oauthService *auth.OAuthService, userService user.Service, jwtSecret string, logger *zap.Logger) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		userService:  userService,
		jwtSecret:    jwtSecret,
		logger:       logger,
	}
}

// GetProviders returns information about available OAuth2 providers
// @Summary Get OAuth providers
// @Description Get list of available OAuth2 providers
// @Tags auth
// @Produce json
// @Success 200 {object} dto.OAuth2ProvidersResponse
// @Failure 500 {object} map[string]string
// @Router /api/auth/oauth/providers [get]
func (h *OAuthHandler) GetProviders(c *gin.Context) {
	providers := h.oauthService.GetProviders()

	response := dto.OAuth2ProvidersResponse{
		Providers: make([]dto.ProviderInfo, 0, len(providers)),
	}

	for name, config := range providers {
		info := dto.ProviderInfo{
			Name:   name,
			Scopes: config.Scopes,
		}

		// Set display name based on provider
		switch name {
		case "google":
			info.DisplayName = "Google"
		case "github":
			info.DisplayName = "GitHub"
		case "facebook":
			info.DisplayName = "Facebook"
		case "twitter":
			info.DisplayName = "Twitter"
		case "microsoft":
			info.DisplayName = "Microsoft"
		default:
			info.DisplayName = name
		}

		response.Providers = append(response.Providers, info)
	}

	c.JSON(http.StatusOK, response)
}

// InitiateLogin starts the OAuth2 flow
// @Summary Initiate OAuth login
// @Description Initiate OAuth2 login flow for a provider
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.OAuth2LoginRequest true "OAuth2 login request"
// @Param provider query string false "Provider name (for GET requests)"
// @Success 200 {object} dto.OAuth2LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/oauth/login [post]
// @Router /api/auth/oauth/login [get]
func (h *OAuthHandler) InitiateLogin(c *gin.Context) {
	var provider string

	// Handle both GET and POST requests
	if c.Request.Method == "GET" {
		// For GET requests, get provider from query parameter
		provider = c.Query("provider")
		if provider == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider parameter is required"})
			return
		}
	} else {
		// For POST requests, get provider from request body
		var req dto.OAuth2LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		provider = req.Provider
	}

	h.logger.Info("Initiating OAuth login", zap.String("provider", provider))

	authURL, state, err := h.oauthService.GetAuthURL(provider)
	if err != nil {
		h.logger.Error("Failed to get auth URL", zap.Error(err), zap.String("provider", provider))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate OAuth login"})
		return
	}

	response := dto.OAuth2LoginResponse{
		AuthURL: authURL,
		State:   state,
	}

	c.JSON(http.StatusOK, response)
}

// HandleCallback processes the OAuth2 callback
// @Summary Handle OAuth callback
// @Description Process the OAuth2 callback after user authorization
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.OAuth2CallbackRequest true "OAuth2 callback data"
// @Param code query string false "Authorization code (for GET requests)"
// @Param state query string false "OAuth state (for GET requests)"
// @Param provider query string false "Provider name (for GET requests)"
// @Success 200 {object} dto.OAuth2CallbackResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/auth/oauth/callback [post]
// @Router /api/auth/oauth/callback [get]
func (h *OAuthHandler) HandleCallback(c *gin.Context) {
	var req dto.OAuth2CallbackRequest

	// Handle both GET and POST requests
	if c.Request.Method == "GET" {
		// For GET requests (typical OAuth redirects), get params from query string
		req.Code = c.Query("code")
		req.State = c.Query("state")
		req.Provider = c.Query("provider")

		// If provider not in query params, try to extract from session or use default
		if req.Provider == "" {
			// Try to extract provider from state
			provider, _ := auth.GetStateStore().GetProviderFromState(req.State)
			if provider != "" {
				req.Provider = provider
			} else {
				// Default to a provider if none specified (must be configured in app)
				req.Provider = "google" // Could be configurable default
			}
		}

		// Validate required parameters
		if req.Code == "" || req.State == "" {
			h.logger.Warn("Missing required parameters in OAuth callback",
				zap.String("code", req.Code),
				zap.String("state", req.State),
				zap.String("provider", req.Provider))
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing required parameters"})
			return
		}
	} else {
		// For POST requests, get params from JSON body
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
	}

	// Log the callback parameters
	h.logger.Info("OAuth callback received",
		zap.String("provider", req.Provider),
		zap.String("code_length", fmt.Sprintf("%d chars", len(req.Code))),
		zap.String("state_length", fmt.Sprintf("%d chars", len(req.State))))

	// Validate state to prevent CSRF
	if !auth.GetStateStore().ValidateState(req.State, req.Provider) {
		h.logger.Warn("Invalid OAuth state", zap.String("provider", req.Provider))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state parameter"})
		return
	}

	// Exchange the authorization code for a token
	token, err := h.oauthService.Exchange(c.Request.Context(), req.Provider, req.Code)
	if err != nil {
		h.logger.Error("Failed to exchange code for token",
			zap.Error(err),
			zap.String("provider", req.Provider))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to authenticate with provider"})
		return
	}

	// Get user info from the provider
	userInfo, err := h.oauthService.GetUserInfo(c.Request.Context(), req.Provider, token)
	if err != nil {
		h.logger.Error("Failed to get user info",
			zap.Error(err),
			zap.String("provider", req.Provider))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user information"})
		return
	}

	// Check if the user exists by provider ID and provider name
	userRecord, err := h.userService.FindUserByProviderID(c.Request.Context(), userInfo.ID, req.Provider)

	if err != nil {
		// User not found, create a new user
		if userInfo.Email == "" {
			h.logger.Error("Provider did not return email", zap.String("provider", req.Provider))
			c.JSON(http.StatusBadRequest, gin.H{"error": "provider did not return email address"})
			return
		}

		// Generate a secure random password since OAuth2 users don't login with password
		password := uuid.New().String()

		createInput := user.CreateUserInput{
			Email:      userInfo.Email,
			Username:   userInfo.Email, // Default to email as username
			Password:   password,
			FirstName:  userInfo.GivenName,
			LastName:   userInfo.FamilyName,
			AvatarURL:  userInfo.Picture,
			ProviderID: userInfo.ID,
			Provider:   req.Provider,
		}

		// If provider didn't return first/last name, use full name
		if createInput.FirstName == "" && createInput.LastName == "" && userInfo.Name != "" {
			createInput.FirstName = userInfo.Name
		}

		// Create the user
		userRecord, err = h.userService.CreateUser(c.Request.Context(), createInput)
		if err != nil {
			h.logger.Error("Failed to create user from OAuth",
				zap.Error(err),
				zap.String("provider", req.Provider))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user account"})
			return
		}
	} else {
		// Update user profile with latest info from provider
		avatarURL := userInfo.Picture
		updateInput := user.UpdateUserInput{
			AvatarURL: &avatarURL,
		}

		// Only update names if they're not set already
		if userRecord.FirstName == "" && userInfo.GivenName != "" {
			firstName := userInfo.GivenName
			updateInput.FirstName = &firstName
		}

		if userRecord.LastName == "" && userInfo.FamilyName != "" {
			lastName := userInfo.FamilyName
			updateInput.LastName = &lastName
		}

		// Update the user
		userRecord, err = h.userService.UpdateUser(c.Request.Context(), userRecord.ID, updateInput)
		if err != nil {
			h.logger.Warn("Failed to update user profile from OAuth",
				zap.Error(err),
				zap.String("userId", userRecord.ID.String()))
			// Continue with login even if update fails
		}
	}

	// Get user's roles and permissions
	roles, permissions, err := h.userService.GetUserRolesAndPermissions(c.Request.Context(), userRecord.ID)
	if err != nil {
		h.logger.Error("Failed to get user roles and permissions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user permissions"})
		return
	}

	// Generate JWT token
	jwtToken, err := auth.GenerateToken(
		userRecord.ID,
		userRecord.Email,
		roles,
		uuid.Nil, // No org ID for now
		permissions,
		h.jwtSecret,
		24, // 24 hour expiry
	)
	if err != nil {
		h.logger.Error("Failed to generate JWT token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// Create session
	session := auth.GetSessionStore().CreateSession(
		userRecord.ID,
		fmt.Sprintf("OAuth via %s", req.Provider),
		c.ClientIP(),
		jwtToken,
		24*time.Hour,
	)

	response := dto.OAuth2CallbackResponse{
		Token:     jwtToken,
		ExpiresAt: session.ExpiresAt.Unix(),
		User: dto.UserResponse{
			ID:          userRecord.ID,
			Email:       userRecord.Email,
			Username:    userRecord.Username,
			FirstName:   userRecord.FirstName,
			LastName:    userRecord.LastName,
			PhoneNumber: userRecord.PhoneNumber,
			AvatarURL:   userRecord.AvatarURL,
			Bio:         userRecord.Bio,
			Timezone:    userRecord.Timezone,
			Locale:      userRecord.Locale,
			IsActive:    userRecord.IsActive,
			CreatedAt:   userRecord.CreatedAt,
			UpdatedAt:   userRecord.UpdatedAt,
		},
	}

	c.JSON(http.StatusOK, response)
}
