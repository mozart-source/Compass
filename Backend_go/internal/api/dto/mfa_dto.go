package dto

// MFASetupResponse represents the MFA setup response
type MFASetupResponse struct {
	Secret       string   `json:"secret"`
	QRCodeBase64 string   `json:"qr_code_base64"`
	OTPAuthURL   string   `json:"otp_auth_url"`
	BackupCodes  []string `json:"backup_codes,omitempty"`
}

// VerifyMFARequest represents the request to verify and enable MFA
type VerifyMFARequest struct {
	Code string `json:"code" binding:"required" example:"123456"`
}

// DisableMFARequest represents the request to disable MFA
type DisableMFARequest struct {
	Password string `json:"password" binding:"required" example:"current-password"`
}

// ValidateMFARequest represents the request to validate an MFA code during login
type ValidateMFARequest struct {
	UserID string `json:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Code   string `json:"code" binding:"required" example:"123456"`
}

// MFAStatusResponse represents the MFA status response
type MFAStatusResponse struct {
	Enabled bool `json:"enabled"`
}

// MFALoginRequest represents the MFA-specific part of a login with 2FA
type MFALoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"securepassword123"`
	Code     string `json:"code,omitempty" example:"123456"` // Optional during first authentication step
}

// MFARequiredResponse indicates that MFA is required to complete login
type MFARequiredResponse struct {
	MFARequired bool   `json:"mfa_required" example:"true"`
	UserID      string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Message     string `json:"message" example:"Please enter your MFA code to complete login"`
	TTL         int    `json:"ttl" example:"300"` // Time-to-live in seconds for the MFA token
}
