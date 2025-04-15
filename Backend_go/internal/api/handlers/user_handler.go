package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/api/dto"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/user"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/security/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

var log = logrus.New()

type UserHandler struct {
	userService user.Service
	jwtSecret   string
}

func NewUserHandler(userService user.Service, jwtSecret string) *UserHandler {
	return &UserHandler{userService: userService, jwtSecret: jwtSecret}
}

// CreateUser handles user registration
// @Summary Create a new user
// @Description Register a new user in the system
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User registration information"
// @Success 201 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/register [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	// Get validated model from context (set by validation middleware)
	validatedModel, exists := c.Get("validated_model")
	var input dto.CreateUserRequest

	if exists {
		// If validation middleware provided the model, use it
		// The model will be a pointer since we created it with reflect.New
		if validatedPtr, ok := validatedModel.(*dto.CreateUserRequest); ok {
			input = *validatedPtr
		} else {
			// Log the actual type for debugging
			log.Errorf("Invalid model type: %T, expected *dto.CreateUserRequest", validatedModel)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid model type from validation"})
			return
		}
	} else {
		// If validation middleware didn't run, do manual binding
		if err := c.ShouldBindJSON(&input); err != nil {
			log.Errorf("Failed to bind CreateUserRequest: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	// Make sure all required fields are present
	if input.Email == "" || input.Username == "" || input.Password == "" ||
		input.FirstName == "" || input.LastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required fields (email, username, password, first_name, last_name)"})
		return
	}

	createInput := user.CreateUserInput{
		Email:       input.Email,
		Username:    input.Username,
		Password:    input.Password,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		PhoneNumber: input.PhoneNumber,
		Timezone:    input.Timezone,
		Locale:      input.Locale,
	}

	createdUser, err := h.userService.CreateUser(c.Request.Context(), createInput)
	if err != nil {
		log.Errorf("Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.UserResponse{
		ID:          createdUser.ID,
		Email:       createdUser.Email,
		Username:    createdUser.Username,
		FirstName:   createdUser.FirstName,
		LastName:    createdUser.LastName,
		PhoneNumber: createdUser.PhoneNumber,
		AvatarURL:   createdUser.AvatarURL,
		Bio:         createdUser.Bio,
		Timezone:    createdUser.Timezone,
		Locale:      createdUser.Locale,
		IsActive:    createdUser.IsActive,
		IsSuperuser: createdUser.IsSuperuser,
		CreatedAt:   createdUser.CreatedAt,
		UpdatedAt:   createdUser.UpdatedAt,
		DeletedAt:   createdUser.DeletedAt,
	}

	c.JSON(http.StatusCreated, gin.H{"user": response})
}

// Login handles user authentication and session creation
// @Summary Login user
// @Description Authenticate user and create a new session
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var loginRequest dto.LoginRequest

	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate user
	user, err := h.userService.AuthenticateUser(c.Request.Context(), loginRequest.Email, loginRequest.Password)
	if err != nil {
		log.Error("Authentication failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Record successful login activity
	activityInput := convertToUserActivityInput(
		user.ID,
		"login_success",
		map[string]interface{}{
			"ip_address": c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
		},
	)
	err = h.userService.RecordUserActivity(c.Request.Context(), activityInput)
	if err != nil {
		log.Error("Failed to record login activity", zap.Error(err))
	}

	// Check if MFA is enabled for the user
	if user.MFAEnabled {
		// Create a temporary auth token for MFA validation - not used now but might be used later
		// Just storing user ID in the response is enough for now
		_, err := auth.GenerateTemporaryToken(
			user.ID,
			user.Email,
			h.jwtSecret,
			5, // 5 minute expiry
		)
		if err != nil {
			log.Error("Failed to generate temporary token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process login"})
			return
		}

		// Return MFA required response
		c.JSON(http.StatusOK, dto.MFARequiredResponse{
			MFARequired: true,
			UserID:      user.ID.String(),
			Message:     "Please enter your MFA code to complete login",
			TTL:         300, // 5 minutes in seconds
		})
		return
	}

	// If MFA not enabled, proceed with normal login flow
	// Get user's roles and permissions
	roles, permissions, err := h.userService.GetUserRolesAndPermissions(c.Request.Context(), user.ID)
	if err != nil {
		log.Error("Failed to get user roles and permissions", zap.Error(err))
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
		log.Error("Failed to generate token", zap.Error(err))
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

	// Record session analytics
	h.recordSessionActivity(c, user.ID, session.ID, "login", session.DeviceInfo, session.IPAddress)

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
			DeletedAt:   user.DeletedAt,
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

// recordSessionActivity is a helper function to record session activities
func (h *UserHandler) recordSessionActivity(c *gin.Context, userID uuid.UUID, sessionID, action, deviceInfo, ipAddress string) {
	input := user.RecordSessionActivityInput{
		SessionID:  sessionID,
		UserID:     userID,
		Action:     action,
		DeviceInfo: deviceInfo,
		IPAddress:  ipAddress,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"user_agent": c.Request.UserAgent(),
			"path":       c.Request.URL.Path,
		},
	}

	err := h.userService.RecordSessionActivity(c.Request.Context(), input)
	if err != nil {
		// Just log the error, don't fail the request
		log.Errorf("Failed to record session activity: %v", err)
	}
}

// GetUser handles fetching a single user by ID
// @Summary Get a user by ID
// @Description Get user details by their ID
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/profile [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	foundUser, err := h.userService.GetUser(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.UserResponse{
		ID:          foundUser.ID,
		Email:       foundUser.Email,
		Username:    foundUser.Username,
		FirstName:   foundUser.FirstName,
		LastName:    foundUser.LastName,
		PhoneNumber: foundUser.PhoneNumber,
		AvatarURL:   foundUser.AvatarURL,
		Bio:         foundUser.Bio,
		Timezone:    foundUser.Timezone,
		Locale:      foundUser.Locale,
		IsActive:    foundUser.IsActive,
		IsSuperuser: foundUser.IsSuperuser,
		CreatedAt:   foundUser.CreatedAt,
		UpdatedAt:   foundUser.UpdatedAt,
		DeletedAt:   foundUser.DeletedAt,
	}

	c.JSON(http.StatusOK, gin.H{"user": response})
}

// UpdateUser handles updating user information
// @Summary Update a user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/profile [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var input dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updateInput := user.UpdateUserInput{
		Username:    input.Username,
		Email:       input.Email,
		FirstName:   input.FirstName,
		LastName:    input.LastName,
		PhoneNumber: input.PhoneNumber,
		AvatarURL:   input.AvatarURL,
		Bio:         input.Bio,
		Timezone:    input.Timezone,
		Locale:      input.Locale,
	}

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), userID.(uuid.UUID), updateInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.UserResponse{
		ID:          updatedUser.ID,
		Email:       updatedUser.Email,
		Username:    updatedUser.Username,
		FirstName:   updatedUser.FirstName,
		LastName:    updatedUser.LastName,
		PhoneNumber: updatedUser.PhoneNumber,
		AvatarURL:   updatedUser.AvatarURL,
		Bio:         updatedUser.Bio,
		Timezone:    updatedUser.Timezone,
		Locale:      updatedUser.Locale,
		IsActive:    updatedUser.IsActive,
		IsSuperuser: updatedUser.IsSuperuser,
		CreatedAt:   updatedUser.CreatedAt,
		UpdatedAt:   updatedUser.UpdatedAt,
		DeletedAt:   updatedUser.DeletedAt,
	}

	c.JSON(http.StatusOK, gin.H{"user": response})
}

// DeleteUser handles user deletion
// @Summary Delete a user
// @Description Delete a user
// @Tags users
// @Accept json
// @Produce json
// @Success 204 "No Content"
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/profile [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	err := h.userService.DeleteUser(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUserRolesAndPermissions retrieves the roles and permissions for a user
func (h *UserHandler) GetUserRolesAndPermissions(c *gin.Context, userID uuid.UUID) ([]string, []string, error) {
	roles, permissions, err := h.userService.GetUserRolesAndPermissions(c.Request.Context(), userID)
	if err != nil {
		return nil, nil, err
	}
	return roles, permissions, nil
}

// Logout handles user logout
// @Summary Logout user
// @Description Invalidate the user's JWT token
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "Successfully logged out"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/users/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	token, exists := c.Get("token")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "no token found"})
		return
	}

	// Get token claims to get expiry time
	claims, err := auth.ValidateToken(token.(string), h.jwtSecret)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Get user ID for analytics
	userID := claims.UserID

	// Get session for analytics
	sessionVal, exists := c.Get("session")
	if exists {
		session := sessionVal.(*auth.Session)
		h.recordSessionActivity(c, userID, session.ID, "logout", session.DeviceInfo, session.IPAddress)
	}

	// Invalidate session
	auth.GetSessionStore().InvalidateSession(token.(string))

	// Add token to blacklist
	auth.GetTokenBlacklist().AddToBlacklist(token.(string), claims.ExpiresAt.Time)

	c.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}

// GetUserSessions returns all active sessions for the current user
// @Summary Get user sessions
// @Description Get all active sessions for the current user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.SessionResponse
// @Failure 401 {object} map[string]string
// @Router /api/users/sessions [get]
func (h *UserHandler) GetUserSessions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	sessions := auth.GetSessionStore().GetUserSessions(userID.(uuid.UUID))

	response := make([]dto.SessionResponse, len(sessions))
	for i, session := range sessions {
		response[i] = dto.SessionResponse{
			ID:           session.ID,
			DeviceInfo:   session.DeviceInfo,
			IPAddress:    session.IPAddress,
			LastActivity: session.LastActivity,
			ExpiresAt:    session.ExpiresAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{"sessions": response})
}

// RevokeSession revokes a specific session
// @Summary Revoke session
// @Description Revoke a specific session by ID
// @Tags users
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/users/sessions/{id}/revoke [post]
func (h *UserHandler) RevokeSession(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	sessionID := c.Param("id")
	sessions := auth.GetSessionStore().GetUserSessions(userID.(uuid.UUID))

	for _, session := range sessions {
		if session.ID == sessionID {
			// Record session revocation in analytics
			h.recordSessionActivity(c, userID.(uuid.UUID), session.ID, "session_revoked", session.DeviceInfo, session.IPAddress)

			auth.GetSessionStore().InvalidateSession(session.Token)
			auth.GetTokenBlacklist().AddToBlacklist(session.Token, session.ExpiresAt)
			c.JSON(http.StatusOK, gin.H{"message": "session revoked successfully"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
}

// GetUserActivity retrieves user activity analytics
// @Summary Get user activity
// @Description Get analytics data for user activities
// @Tags users
// @Accept json
// @Produce json
// @Param filter query dto.UserAnalyticsFilter false "Filter parameters"
// @Success 200 {object} dto.UserAnalyticsListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/analytics/activity [get]
func (h *UserHandler) GetUserActivity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var filter dto.UserAnalyticsFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse time strings to time.Time
	startTime, err := time.Parse(time.RFC3339, filter.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, expected RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, filter.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, expected RFC3339"})
		return
	}

	analytics, total, err := h.userService.GetUserAnalytics(
		c.Request.Context(),
		userID.(uuid.UUID),
		startTime,
		endTime,
		filter.Page,
		filter.PageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert domain entities to DTO responses
	responseItems := make([]dto.UserAnalyticsResponse, len(analytics))
	for i, item := range analytics {
		// Parse metadata JSON
		var metadata map[string]interface{}
		if item.Metadata != "" {
			if err := json.Unmarshal([]byte(item.Metadata), &metadata); err != nil {
				// Just use empty metadata if parsing fails
				metadata = make(map[string]interface{})
			}
		} else {
			metadata = make(map[string]interface{})
		}

		responseItems[i] = dto.UserAnalyticsResponse{
			ID:        item.ID,
			UserID:    item.UserID,
			Action:    item.Action,
			Timestamp: item.Timestamp,
			Metadata:  metadata,
		}
	}

	response := dto.UserAnalyticsListResponse{
		Analytics:  responseItems,
		TotalCount: total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetSessionActivity retrieves session activity analytics
// @Summary Get session activity
// @Description Get analytics data for session activities
// @Tags users
// @Accept json
// @Produce json
// @Param filter query dto.UserAnalyticsFilter false "Filter parameters"
// @Success 200 {object} dto.SessionAnalyticsListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/analytics/sessions [get]
func (h *UserHandler) GetSessionActivity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var filter dto.UserAnalyticsFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse time strings to time.Time
	startTime, err := time.Parse(time.RFC3339, filter.StartTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, expected RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, filter.EndTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, expected RFC3339"})
		return
	}

	analytics, total, err := h.userService.GetSessionAnalytics(
		c.Request.Context(),
		userID.(uuid.UUID),
		startTime,
		endTime,
		filter.Page,
		filter.PageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert domain entities to DTO responses
	responseItems := make([]dto.SessionAnalyticsResponse, len(analytics))
	for i, item := range analytics {
		// Parse metadata JSON
		var metadata map[string]interface{}
		if item.Metadata != "" {
			if err := json.Unmarshal([]byte(item.Metadata), &metadata); err != nil {
				// Just use empty metadata if parsing fails
				metadata = make(map[string]interface{})
			}
		} else {
			metadata = make(map[string]interface{})
		}

		responseItems[i] = dto.SessionAnalyticsResponse{
			ID:         item.ID,
			SessionID:  item.SessionID,
			UserID:     item.UserID,
			Action:     item.Action,
			DeviceInfo: item.DeviceInfo,
			IPAddress:  item.IPAddress,
			Timestamp:  item.Timestamp,
			Metadata:   metadata,
		}
	}

	response := dto.SessionAnalyticsListResponse{
		Analytics:  responseItems,
		TotalCount: total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// GetUserActivitySummary retrieves a summary of user activity
// @Summary Get user activity summary
// @Description Get a summary of user activity counts by action type
// @Tags users
// @Accept json
// @Produce json
// @Param start_time query string true "Start time (RFC3339)" format(date-time)
// @Param end_time query string true "End time (RFC3339)" format(date-time)
// @Success 200 {object} dto.UserActivitySummaryResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/analytics/summary [get]
func (h *UserHandler) GetUserActivitySummary(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time and end_time are required"})
		return
	}

	// Parse time strings to time.Time
	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format, expected RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_time format, expected RFC3339"})
		return
	}

	summary, err := h.userService.GetUserActivitySummary(
		c.Request.Context(),
		userID.(uuid.UUID),
		startTime,
		endTime,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := dto.UserActivitySummaryResponse{
		UserID:       summary.UserID,
		ActionCounts: summary.ActionCounts,
		StartTime:    summary.StartTime,
		EndTime:      summary.EndTime,
		TotalActions: summary.TotalActions,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// RecordUserActivity handles manual recording of user activity
// @Summary Record user activity
// @Description Manually record a user activity for analytics
// @Tags users
// @Accept json
// @Produce json
// @Param activity body dto.RecordUserActivityRequest true "Activity details"
// @Success 201 "Activity recorded successfully"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/analytics/record [post]
func (h *UserHandler) RecordUserActivity(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var request dto.RecordUserActivityRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := user.RecordUserActivityInput{
		UserID:    userID.(uuid.UUID),
		Action:    request.Action,
		Metadata:  request.Metadata,
		Timestamp: time.Now(),
	}

	if err := h.userService.RecordUserActivity(c.Request.Context(), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusCreated)
}

// convertToUserActivityInput converts the activity input to the domain type
func convertToUserActivityInput(userID uuid.UUID, action string, metadata map[string]interface{}) user.RecordUserActivityInput {
	return user.RecordUserActivityInput{
		UserID:    userID,
		Action:    action,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
}
