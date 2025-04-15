package mfa

import (
	"bytes"
	"crypto/rand"
	"encoding/base32"
	"image/png"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// TOTPConfig holds configuration for TOTP generation
type TOTPConfig struct {
	Issuer      string // The provider/company name
	AccountName string // Usually the user's email or username
	SecretSize  int    // Size of the TOTP secret
	Digits      otp.Digits
	Algorithm   otp.Algorithm
	Period      uint
}

// DefaultTOTPConfig returns a default configuration for TOTP
func DefaultTOTPConfig(accountName string) TOTPConfig {
	return TOTPConfig{
		Issuer:      "Compass",
		AccountName: accountName,
		SecretSize:  20,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
		Period:      30,
	}
}

// GenerateSecret generates a new TOTP secret
func GenerateSecret(config TOTPConfig) (string, error) {
	// Generate random bytes
	secret := make([]byte, config.SecretSize)
	_, err := rand.Read(secret)
	if err != nil {
		return "", err
	}

	// Encode as base32 (RFC 4648) and remove padding
	encodedSecret := base32.StdEncoding.EncodeToString(secret)
	encodedSecret = strings.TrimRight(encodedSecret, "=")
	return encodedSecret, nil
}

// GenerateTOTPKey generates a new TOTP key
func GenerateTOTPKey(config TOTPConfig) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      config.Issuer,
		AccountName: config.AccountName,
		SecretSize:  uint(config.SecretSize),
		Digits:      config.Digits,
		Algorithm:   config.Algorithm,
		Period:      config.Period,
	})
}

// ValidateTOTP validates a TOTP code against a secret
func ValidateTOTP(code string, secret string) bool {
	// Use the provided totp library to validate the code
	return totp.Validate(code, secret)
}

// GenerateQRCode generates a QR code PNG image for a TOTP key
func GenerateQRCode(key *otp.Key) ([]byte, error) {
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, err
	}

	err = png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GetTOTPCode generates a current TOTP code for a secret
func GetTOTPCode(secret string) (string, error) {
	return totp.GenerateCode(secret, time.Now())
}
