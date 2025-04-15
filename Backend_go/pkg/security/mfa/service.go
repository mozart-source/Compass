package mfa

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/pquerna/otp"
)

// Common errors
var (
	ErrGeneratingSecret = errors.New("error generating TOTP secret")
	ErrInvalidCode      = errors.New("invalid verification code")
	ErrExpiredCode      = errors.New("verification code expired")
)

// Service provides MFA functionality
type Service interface {
	// Setup generates a new TOTP secret for a user
	Setup(accountName string) (*SetupResult, error)

	// Validate checks if a provided TOTP code is valid for the given secret
	Validate(secret, code string) (bool, error)

	// GenerateBackupCodes generates a set of one-time backup codes
	GenerateBackupCodes() ([]string, error)
}

// SetupResult contains the result of setting up MFA
type SetupResult struct {
	Secret    string // The TOTP secret
	QRCode    []byte // QR code as PNG image bytes
	URI       string // otpauth URI for manual entry
	CreatedAt time.Time
}

type service struct {
	config TOTPConfig
}

// NewService creates a new MFA service
func NewService(issuer string) Service {
	return &service{
		config: TOTPConfig{
			Issuer:     issuer,
			SecretSize: 20,
			Digits:     otp.DigitsSix,
			Algorithm:  otp.AlgorithmSHA1,
			Period:     30,
		},
	}
}

// Setup implements the Service interface
func (s *service) Setup(accountName string) (*SetupResult, error) {
	// Set account name in config
	config := s.config
	config.AccountName = accountName

	// Generate a TOTP key
	key, err := GenerateTOTPKey(config)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrGeneratingSecret, err)
	}

	// Generate QR code
	qrCode, err := GenerateQRCode(key)
	if err != nil {
		return nil, fmt.Errorf("error generating QR code: %v", err)
	}

	return &SetupResult{
		Secret:    key.Secret(),
		QRCode:    qrCode,
		URI:       key.URL(),
		CreatedAt: time.Now(),
	}, nil
}

// Validate implements the Service interface
func (s *service) Validate(secret, code string) (bool, error) {
	// Basic validation
	if secret == "" || code == "" {
		return false, fmt.Errorf("secret and code cannot be empty")
	}

	// Validate the TOTP code
	valid := ValidateTOTP(code, secret)
	if !valid {
		return false, ErrInvalidCode
	}

	return true, nil
}

// GenerateBackupCodes creates a set of backup codes for recovery purposes
func (s *service) GenerateBackupCodes() ([]string, error) {
	const (
		numCodes    = 10
		codeLength  = 8
		charsPerInt = 4
	)

	// Generate backup codes
	codes := make([]string, numCodes)
	for i := 0; i < numCodes; i++ {
		// Generate 8 random bytes
		randomBytes := make([]byte, codeLength)
		_, err := rand.Read(randomBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to generate random bytes: %v", err)
		}

		// Convert to a numeric code (use modulo to get numbers 0-9)
		code := ""
		for j := 0; j < codeLength; j++ {
			// Get a number 0-9 from random byte
			digit := int(randomBytes[j]) % 10
			code += fmt.Sprintf("%d", digit)
		}

		// Format code with a hyphen in the middle (e.g., 1234-5678)
		codes[i] = fmt.Sprintf("%s-%s", code[:charsPerInt], code[charsPerInt:])
	}

	return codes, nil
}
