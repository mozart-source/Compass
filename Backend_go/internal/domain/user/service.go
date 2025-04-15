package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"encoding/json"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/events"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/domain/roles"
	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/cache"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/security/mfa"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var log = logrus.New()

// Input types
type CreateUserInput struct {
	Email        string                 `json:"email"`
	Username     string                 `json:"username"`
	Password     string                 `json:"password"`
	FirstName    string                 `json:"first_name"`
	LastName     string                 `json:"last_name"`
	PhoneNumber  string                 `json:"phone_number,omitempty"`
	AvatarURL    string                 `json:"avatar_url,omitempty"`
	Bio          string                 `json:"bio,omitempty"`
	Timezone     string                 `json:"timezone,omitempty"`
	Locale       string                 `json:"locale,omitempty"`
	Preferences  map[string]interface{} `json:"preferences,omitempty"`
	Provider     string                 `json:"provider,omitempty"`
	ProviderID   string                 `json:"provider_id,omitempty"`
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
}

type UpdateUserInput struct {
	Email        *string                `json:"email,omitempty"`
	Username     *string                `json:"username,omitempty"`
	Password     *string                `json:"password,omitempty"`
	FirstName    *string                `json:"first_name,omitempty"`
	LastName     *string                `json:"last_name,omitempty"`
	PhoneNumber  *string                `json:"phone_number,omitempty"`
	AvatarURL    *string                `json:"avatar_url,omitempty"`
	Bio          *string                `json:"bio,omitempty"`
	Timezone     *string                `json:"timezone,omitempty"`
	Locale       *string                `json:"locale,omitempty"`
	Preferences  map[string]interface{} `json:"preferences,omitempty"`
	Provider     *string                `json:"provider,omitempty"`
	ProviderID   *string                `json:"provider_id,omitempty"`
	ProviderData map[string]interface{} `json:"provider_data,omitempty"`
}

// Analytics types
type RecordUserActivityInput struct {
	UserID    uuid.UUID              `json:"user_id"`
	Action    string                 `json:"action"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
}

type RecordSessionActivityInput struct {
	SessionID  string                 `json:"session_id"`
	UserID     uuid.UUID              `json:"user_id"`
	Action     string                 `json:"action"`
	DeviceInfo string                 `json:"device_info,omitempty"`
	IPAddress  string                 `json:"ip_address,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp,omitempty"`
}

type UserActivitySummary struct {
	UserID       uuid.UUID      `json:"user_id"`
	ActionCounts map[string]int `json:"action_counts"`
	StartTime    time.Time      `json:"start_time"`
	EndTime      time.Time      `json:"end_time"`
	TotalActions int            `json:"total_actions"`
}

// Common errors
var (
	ErrEmailExists        = errors.New("email already exists")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account is locked")
	ErrAccountInactive    = errors.New("account is inactive")
)

// MFASetupResponse represents the response for MFA setup
type MFASetupResponse struct {
	Secret       string   `json:"secret"`
	QRCodeBase64 string   `json:"qr_code_base64"`
	OTPAuthURL   string   `json:"otp_auth_url"`
	BackupCodes  []string `json:"backup_codes,omitempty"`
}

// Define UserDashboardMetrics struct for dashboard metrics aggregation
// UserDashboardMetrics represents summary metrics for the dashboard
// Used by GetDashboardMetrics
type UserDashboardMetrics struct {
	ActivitySummary map[string]int
}

// Service interface
type Service interface {
	CreateUser(ctx context.Context, input CreateUserInput) (*User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	FindUserByProviderID(ctx context.Context, providerID, provider string) (*User, error)
	ListUsers(ctx context.Context, filter UserFilter) ([]User, int64, error)
	UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	AuthenticateUser(ctx context.Context, email, password string) (*User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error
	LockAccount(ctx context.Context, id uuid.UUID, duration time.Duration) error
	UnlockAccount(ctx context.Context, id uuid.UUID) error
	GetUserRolesAndPermissions(ctx context.Context, userID uuid.UUID) ([]string, []string, error)

	// Analytics methods
	RecordUserActivity(ctx context.Context, input RecordUserActivityInput) error
	RecordSessionActivity(ctx context.Context, input RecordSessionActivityInput) error
	GetUserAnalytics(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]UserAnalytics, int64, error)
	GetSessionAnalytics(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]SessionAnalytics, int64, error)
	GetUserActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (*UserActivitySummary, error)

	// MFA methods
	SetupMFA(ctx context.Context, userID uuid.UUID) (*MFASetupResponse, error)
	VerifyAndEnableMFA(ctx context.Context, userID uuid.UUID, code string) error
	ValidateMFACode(ctx context.Context, userID uuid.UUID, code string) (bool, error)
	DisableMFA(ctx context.Context, userID uuid.UUID, password string) error
	IsMFAEnabled(ctx context.Context, userID uuid.UUID) (bool, error)

	// New method
	GetDashboardMetrics(userID uuid.UUID) (UserDashboardMetrics, error)
}

type service struct {
	repo         Repository
	rolesService roles.Service
	mfaService   mfa.Service
	redis        *cache.RedisClient
}

func NewService(repo Repository, rolesService roles.Service, redis *cache.RedisClient) Service {
	return &service{
		repo:         repo,
		rolesService: rolesService,
		mfaService:   mfa.NewService("Compass"),
		redis:        redis,
	}
}

// validateCreateUserInput validates the input for creating a user
func validateCreateUserInput(input CreateUserInput) error {
	if input.Email == "" {
		return errors.New("email is required")
	}
	if input.Username == "" {
		return errors.New("username is required")
	}
	if input.Password == "" {
		return errors.New("password is required")
	}
	if input.FirstName == "" {
		return errors.New("first name is required")
	}
	if input.LastName == "" {
		return errors.New("last name is required")
	}
	return nil
}

// Helper to marshal metadata
func marshalMetadata(data map[string]interface{}) string {
	b, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// CreateUser creates a new user with the given input
func (s *service) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	if err := validateCreateUserInput(input); err != nil {
		return nil, err
	}

	// Check if email already exists
	existingUser, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, fmt.Errorf("checking email existence: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	// Check if username already exists
	existingUser, err = s.repo.FindByUsername(ctx, input.Username)
	if err != nil && !errors.Is(err, ErrUserNotFound) {
		return nil, fmt.Errorf("checking username existence: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUsernameExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	// Create user
	user := &User{
		ID:           uuid.New(),
		Email:        input.Email,
		Username:     input.Username,
		PasswordHash: string(hashedPassword),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		PhoneNumber:  input.PhoneNumber,
		AvatarURL:    input.AvatarURL,
		Bio:          input.Bio,
		Timezone:     input.Timezone,
		Locale:       input.Locale,
		Status:       UserStatusActive,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Preferences:  input.Preferences,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	// Get default user role
	defaultRole, err := s.rolesService.GetRoleByName(ctx, "user")
	if err != nil {
		return nil, fmt.Errorf("getting default role: %w", err)
	}

	// Assign default role to user
	if err := s.rolesService.AssignRoleToUser(ctx, user.ID, defaultRole.ID); err != nil {
		return nil, fmt.Errorf("assigning default role: %w", err)
	}

	// Record user creation activity
	s.recordUserRegistration(ctx, user.ID)

	return user, nil
}

func (s *service) recordUserRegistration(ctx context.Context, userID uuid.UUID) {
	analytics := &UserAnalytics{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    "user_registration",
		Timestamp: time.Now(),
		Metadata:  `{"source": "direct"}`,
	}

	// We don't want analytics recording to fail user creation, so just log any errors
	if err := s.repo.RecordUserActivity(ctx, analytics); err != nil {
		// In a real application, log this error
		fmt.Printf("Error recording user registration analytics: %v\n", err)
	}
}

func (s *service) GetUser(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *service) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	return s.repo.FindByEmail(ctx, email)
}

func (s *service) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	return s.repo.FindByUsername(ctx, username)
}

func (s *service) FindUserByProviderID(ctx context.Context, providerID, provider string) (*User, error) {
	return s.repo.FindByProviderID(ctx, providerID, provider)
}

func (s *service) ListUsers(ctx context.Context, filter UserFilter) ([]User, int64, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *service) UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	oldEmail := user.Email
	oldUsername := user.Username

	// Update fields if provided
	if input.Email != nil {
		// Check if new email exists
		if *input.Email != user.Email {
			existingUser, err := s.repo.FindByEmail(ctx, *input.Email)
			if err != nil {
				return nil, err
			}
			if existingUser != nil {
				return nil, ErrEmailExists
			}
		}
		user.Email = *input.Email
	}

	if input.Username != nil {
		user.Username = *input.Username
	}

	if input.Password != nil {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = string(hashedPassword)
	}

	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}

	if input.LastName != nil {
		user.LastName = *input.LastName
	}

	if input.PhoneNumber != nil {
		user.PhoneNumber = *input.PhoneNumber
	}

	if input.AvatarURL != nil {
		user.AvatarURL = *input.AvatarURL
	}

	if input.Bio != nil {
		user.Bio = *input.Bio
	}

	if input.Timezone != nil {
		user.Timezone = *input.Timezone
	}

	if input.Locale != nil {
		user.Locale = *input.Locale
	}

	if input.Preferences != nil {
		user.Preferences = input.Preferences
	}

	user.UpdatedAt = time.Now()
	err = s.repo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	// Record profile update analytics
	s.recordProfileUpdate(ctx, user.ID)

	// Record email/username change analytics
	if input.Email != nil && *input.Email != oldEmail {
		metadata := marshalMetadata(map[string]interface{}{
			"old_email": oldEmail,
			"new_email": *input.Email,
		})
		analytics := &UserAnalytics{
			ID:        uuid.New(),
			UserID:    user.ID,
			Action:    "email_changed",
			Timestamp: time.Now(),
			Metadata:  metadata,
		}
		_ = s.repo.RecordUserActivity(ctx, analytics)
	}
	if input.Username != nil && *input.Username != oldUsername {
		metadata := marshalMetadata(map[string]interface{}{
			"old_username": oldUsername,
			"new_username": *input.Username,
		})
		analytics := &UserAnalytics{
			ID:        uuid.New(),
			UserID:    user.ID,
			Action:    "username_changed",
			Timestamp: time.Now(),
			Metadata:  metadata,
		}
		_ = s.repo.RecordUserActivity(ctx, analytics)
	}

	s.recordUserActivity(ctx, user.ID, "profile_updated", map[string]interface{}{
		"updated_fields": getUpdatedFields(input),
	})
	return user, nil
}

func (s *service) recordProfileUpdate(ctx context.Context, userID uuid.UUID) {
	analytics := &UserAnalytics{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    "profile_update",
		Timestamp: time.Now(),
	}

	// Non-blocking analytics recording
	if err := s.repo.RecordUserActivity(ctx, analytics); err != nil {
		// Log error in a real application
		fmt.Printf("Error recording profile update analytics: %v\n", err)
	}
}

func (s *service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Soft delete
	now := time.Now()
	user.DeletedAt = &now
	user.IsActive = false
	err = s.repo.Update(ctx, user)
	if err != nil {
		return err
	}

	// Record user deletion analytics
	if callerID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		metadata := marshalMetadata(map[string]interface{}{
			"deleted_by": callerID.String(),
			"user_id":    user.ID.String(),
			"email":      user.Email,
		})
		analytics := &UserAnalytics{
			ID:        uuid.New(),
			UserID:    user.ID,
			Action:    "user_deleted",
			Timestamp: time.Now(),
			Metadata:  metadata,
		}
		_ = s.repo.RecordUserActivity(ctx, analytics)
	}

	return nil
}

func (s *service) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	if user.AccountLockedUntil != nil && user.AccountLockedUntil.After(time.Now()) {
		return nil, ErrAccountLocked
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.recordUserActivity(ctx, user.ID, "login_failed", nil)
		return nil, ErrInvalidCredentials
	}

	s.recordUserActivity(ctx, user.ID, "login_success", nil)
	return user, nil
}

func (s *service) recordLoginAttempt(ctx context.Context, userID uuid.UUID, success bool) {
	action := "login_success"
	if !success {
		action = "login_failure"
	}

	// Create proper JSON metadata
	metadata := marshalMetadata(map[string]interface{}{
		"success": success,
		"type":    "login",
	})

	analytics := &UserAnalytics{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    action,
		Timestamp: time.Now(),
		Metadata:  metadata,
	}

	if err := s.repo.RecordUserActivity(ctx, analytics); err != nil {
		// Log error in a real application
		fmt.Printf("Error recording login attempt analytics: %v\n", err)
	}
}

func (s *service) UpdatePassword(ctx context.Context, id uuid.UUID, currentPassword, newPassword string) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, user)
	if err != nil {
		return err
	}

	// Record password change
	s.recordPasswordChange(ctx, user.ID)

	return nil
}

func (s *service) recordPasswordChange(ctx context.Context, userID uuid.UUID) {
	analytics := &UserAnalytics{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    "password_change",
		Timestamp: time.Now(),
	}

	if err := s.repo.RecordUserActivity(ctx, analytics); err != nil {
		// Log error in a real application
		fmt.Printf("Error recording password change analytics: %v\n", err)
	}
}

func (s *service) LockAccount(ctx context.Context, id uuid.UUID, duration time.Duration) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	lockUntil := time.Now().Add(duration)
	user.AccountLockedUntil = &lockUntil
	user.UpdatedAt = time.Now()
	err = s.repo.Update(ctx, user)
	if err != nil {
		return err
	}

	// Record account lock analytics
	if callerID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		metadata := marshalMetadata(map[string]interface{}{
			"locked_by": callerID.String(),
			"user_id":   user.ID.String(),
			"duration":  duration.String(),
		})
		analytics := &UserAnalytics{
			ID:        uuid.New(),
			UserID:    user.ID,
			Action:    "account_locked",
			Timestamp: time.Now(),
			Metadata:  metadata,
		}
		_ = s.repo.RecordUserActivity(ctx, analytics)
	}

	return nil
}

func (s *service) UnlockAccount(ctx context.Context, id uuid.UUID) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	user.AccountLockedUntil = nil
	user.FailedLoginAttempts = 0
	user.UpdatedAt = time.Now()
	err = s.repo.Update(ctx, user)
	if err != nil {
		return err
	}

	// Record account unlock analytics
	if callerID, ok := ctx.Value("user_id").(uuid.UUID); ok {
		metadata := marshalMetadata(map[string]interface{}{
			"unlocked_by": callerID.String(),
			"user_id":     user.ID.String(),
		})
		analytics := &UserAnalytics{
			ID:        uuid.New(),
			UserID:    user.ID,
			Action:    "account_unlocked",
			Timestamp: time.Now(),
			Metadata:  metadata,
		}
		_ = s.repo.RecordUserActivity(ctx, analytics)
	}

	return nil
}

// GetUserRolesAndPermissions retrieves the roles and permissions for a given user
func (s *service) GetUserRolesAndPermissions(ctx context.Context, userID uuid.UUID) ([]string, []string, error) {
	// Get user roles
	roles, err := s.rolesService.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting user roles: %w", err)
	}

	// Get user permissions
	permissions, err := s.rolesService.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("getting user permissions: %w", err)
	}

	// Convert roles and permissions to string arrays
	roleNames := make([]string, len(roles))
	for i, role := range roles {
		roleNames[i] = role.Name
	}

	permissionNames := make([]string, len(permissions))
	for i, permission := range permissions {
		permissionNames[i] = permission.Name
	}

	return roleNames, permissionNames, nil
}

// Analytics implementation
func (s *service) RecordUserActivity(ctx context.Context, input RecordUserActivityInput) error {
	timestamp := input.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	metadata := ""
	if input.Metadata != nil {
		// In a real implementation, we'd properly marshal this to JSON
		// For simplicity, we'll use a placeholder string
		metadata = `{"data": "sample"}`
	}

	analytics := &UserAnalytics{
		ID:        uuid.New(),
		UserID:    input.UserID,
		Action:    input.Action,
		Timestamp: timestamp,
		Metadata:  metadata,
	}

	if err := s.repo.RecordUserActivity(ctx, analytics); err != nil {
		return err
	}

	// Publish dashboard event
	event := &events.DashboardEvent{
		EventType: events.EventTypeUserActivity,
		UserID:    input.UserID,
		Timestamp: time.Now().UTC(),
		Details:   input,
	}
	if err := s.redis.PublishDashboardEvent(ctx, event); err != nil {
		log.Error("Failed to publish dashboard event", zap.Error(err))
	}

	return nil
}

func (s *service) RecordSessionActivity(ctx context.Context, input RecordSessionActivityInput) error {
	timestamp := input.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	metadata := ""
	if input.Metadata != nil {
		// In a real implementation, we'd properly marshal this to JSON
		// For simplicity, we'll use a placeholder string
		metadata = `{"data": "sample"}`
	}

	analytics := &SessionAnalytics{
		ID:         uuid.New(),
		SessionID:  input.SessionID,
		UserID:     input.UserID,
		Action:     input.Action,
		DeviceInfo: input.DeviceInfo,
		IPAddress:  input.IPAddress,
		Timestamp:  timestamp,
		Metadata:   metadata,
	}

	return s.repo.RecordSessionActivity(ctx, analytics)
}

func (s *service) GetUserAnalytics(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]UserAnalytics, int64, error) {
	filter := AnalyticsFilter{
		UserID:    &userID,
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      page,
		PageSize:  pageSize,
	}

	return s.repo.GetUserAnalytics(ctx, filter)
}

func (s *service) GetSessionAnalytics(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time, page, pageSize int) ([]SessionAnalytics, int64, error) {
	filter := AnalyticsFilter{
		UserID:    &userID,
		StartTime: &startTime,
		EndTime:   &endTime,
		Page:      page,
		PageSize:  pageSize,
	}

	return s.repo.GetSessionAnalytics(ctx, filter)
}

func (s *service) GetUserActivitySummary(ctx context.Context, userID uuid.UUID, startTime, endTime time.Time) (*UserActivitySummary, error) {
	actionCounts, err := s.repo.GetUserActivitySummary(ctx, userID, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Calculate total actions
	totalActions := 0
	for _, count := range actionCounts {
		totalActions += count
	}

	return &UserActivitySummary{
		UserID:       userID,
		ActionCounts: actionCounts,
		StartTime:    startTime,
		EndTime:      endTime,
		TotalActions: totalActions,
	}, nil
}

// Add a stub for role/permission change analytics (to be called from wherever roles/permissions are changed)
func (s *service) RecordRoleChange(ctx context.Context, userID, changedBy uuid.UUID, oldRoles, newRoles []string) {
	metadata := marshalMetadata(map[string]interface{}{
		"changed_by": changedBy.String(),
		"user_id":    userID.String(),
		"old_roles":  oldRoles,
		"new_roles":  newRoles,
	})
	analytics := &UserAnalytics{
		ID:        uuid.New(),
		UserID:    userID,
		Action:    "roles_changed",
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
	_ = s.repo.RecordUserActivity(ctx, analytics)
}

// SetupMFA generates a new MFA secret for a user
func (s *service) SetupMFA(ctx context.Context, userID uuid.UUID) (*MFASetupResponse, error) {
	// Get the user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Generate TOTP setup
	setupResult, err := s.mfaService.Setup(user.Email)
	if err != nil {
		return nil, fmt.Errorf("error setting up MFA: %w", err)
	}

	// Generate backup codes
	backupCodes, err := s.mfaService.GenerateBackupCodes()
	if err != nil {
		return nil, fmt.Errorf("error generating backup codes: %w", err)
	}

	// Store secret temporarily (not enabled yet)
	user.MFASecret = setupResult.Secret

	// Store backup codes as JSON
	backupCodesJSON, err := json.Marshal(backupCodes)
	if err != nil {
		return nil, fmt.Errorf("error serializing backup codes: %w", err)
	}
	user.MFABackupCodesHash = string(backupCodesJSON)

	user.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("error updating user with MFA secret: %w", err)
	}

	// Convert QR code to base64
	qrCodeBase64 := ""
	if setupResult.QRCode != nil {
		qrCodeBase64 = string(setupResult.QRCode)
	}

	// Return setup info
	return &MFASetupResponse{
		Secret:       setupResult.Secret,
		QRCodeBase64: qrCodeBase64,
		OTPAuthURL:   setupResult.URI,
		BackupCodes:  backupCodes,
	}, nil
}

// VerifyAndEnableMFA verifies a TOTP code and enables MFA for a user
func (s *service) VerifyAndEnableMFA(ctx context.Context, userID uuid.UUID, code string) error {
	// Get the user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify the TOTP code
	valid, err := s.mfaService.Validate(user.MFASecret, code)
	if err != nil {
		return fmt.Errorf("error validating TOTP code: %w", err)
	}
	if !valid {
		return mfa.ErrInvalidCode
	}

	// Enable MFA
	user.MFAEnabled = true
	user.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("error enabling MFA: %w", err)
	}

	// Record MFA enabled activity
	metadata := marshalMetadata(map[string]interface{}{
		"action": "mfa_enabled",
	})
	analytics := &UserAnalytics{
		ID:        uuid.New(),
		UserID:    user.ID,
		Action:    "security_settings_changed",
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
	_ = s.repo.RecordUserActivity(ctx, analytics)

	return nil
}

// ValidateMFACode validates a TOTP code for a user
func (s *service) ValidateMFACode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	// Get the user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, ErrUserNotFound
	}

	// Check if MFA is enabled
	if !user.MFAEnabled {
		return false, errors.New("MFA not enabled for this user")
	}

	// First, try to validate as a TOTP code
	valid, err := s.mfaService.Validate(user.MFASecret, code)
	if valid && err == nil {
		return true, nil
	}

	// If not a valid TOTP code, check if it's a backup code
	var backupCodes []string
	if user.MFABackupCodesHash != "" {
		if err := json.Unmarshal([]byte(user.MFABackupCodesHash), &backupCodes); err == nil {
			// Check if the provided code matches any backup code
			for i, backupCode := range backupCodes {
				if backupCode == code {
					// Remove the used backup code
					backupCodes = append(backupCodes[:i], backupCodes[i+1:]...)

					// Update backup codes in the database
					updatedCodesJSON, err := json.Marshal(backupCodes)
					if err == nil {
						user.MFABackupCodesHash = string(updatedCodesJSON)
						_ = s.repo.Update(ctx, user) // Best effort update
					}

					return true, nil
				}
			}
		}
	}

	// If we reach here, neither TOTP nor backup code was valid
	return false, mfa.ErrInvalidCode
}

// DisableMFA disables MFA for a user
func (s *service) DisableMFA(ctx context.Context, userID uuid.UUID, password string) error {
	// Get the user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	// Verify the password for security
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Disable MFA
	user.MFAEnabled = false
	user.MFASecret = ""
	user.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, user); err != nil {
		return fmt.Errorf("error disabling MFA: %w", err)
	}

	// Record MFA disabled activity
	metadata := marshalMetadata(map[string]interface{}{
		"action": "mfa_disabled",
	})
	analytics := &UserAnalytics{
		ID:        uuid.New(),
		UserID:    user.ID,
		Action:    "security_settings_changed",
		Timestamp: time.Now(),
		Metadata:  metadata,
	}
	_ = s.repo.RecordUserActivity(ctx, analytics)

	return nil
}

// IsMFAEnabled checks if MFA is enabled for a user
func (s *service) IsMFAEnabled(ctx context.Context, userID uuid.UUID) (bool, error) {
	// Get the user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, ErrUserNotFound
	}

	return user.MFAEnabled, nil
}

// GetDashboardMetrics returns dashboard metrics for a user
func (s *service) GetDashboardMetrics(userID uuid.UUID) (UserDashboardMetrics, error) {
	ctx := context.Background()

	// Get analytics for the last 30 days
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	// Get user activity summary
	summary, err := s.repo.GetUserActivitySummary(ctx, userID, startTime, endTime)
	if err != nil {
		return UserDashboardMetrics{}, err
	}

	// Initialize action counts
	actionCounts := map[string]int{
		"actions": 0,
		"logins":  0,
	}

	// Count total actions and logins
	for action, count := range summary {
		actionCounts["actions"] += count
		if action == "login_success" {
			actionCounts["logins"] = count
		}
	}

	metrics := UserDashboardMetrics{
		ActivitySummary: actionCounts,
	}

	// Publish dashboard event
	event := &events.DashboardEvent{
		EventType: events.EventTypeDashboardUpdate,
		UserID:    userID,
		Timestamp: time.Now().UTC(),
		Details:   metrics,
	}
	if err := s.redis.PublishDashboardEvent(ctx, event); err != nil {
		log.Error("Failed to publish dashboard event", zap.Error(err))
	}

	return metrics, nil
}

func (s *service) recordUserActivity(ctx context.Context, userID uuid.UUID, action string, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["action"] = action

	// Publish dashboard event for cache invalidation
	event := &events.DashboardEvent{
		EventType: events.DashboardEventCacheInvalidate,
		UserID:    userID,
		EntityID:  userID,
		Timestamp: time.Now().UTC(),
		Details:   metadata,
	}
	if err := s.redis.PublishDashboardEvent(ctx, event); err != nil {
		zap.L().Error("Failed to publish dashboard event", zap.Error(err))
	}
}

func getUpdatedFields(input UpdateUserInput) []string {
	var fields []string
	if input.Email != nil {
		fields = append(fields, "email")
	}
	if input.Username != nil {
		fields = append(fields, "username")
	}
	if input.FirstName != nil {
		fields = append(fields, "first_name")
	}
	if input.LastName != nil {
		fields = append(fields, "last_name")
	}
	if input.PhoneNumber != nil {
		fields = append(fields, "phone_number")
	}
	if input.AvatarURL != nil {
		fields = append(fields, "avatar_url")
	}
	if input.Bio != nil {
		fields = append(fields, "bio")
	}
	if input.Timezone != nil {
		fields = append(fields, "timezone")
	}
	if input.Locale != nil {
		fields = append(fields, "locale")
	}
	if input.Preferences != nil {
		fields = append(fields, "preferences")
	}
	return fields
}
