package auth

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Custom claims structure
type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	OrgID       uuid.UUID `json:"org_id"`
	Permissions []string  `json:"permissions"`
	jwt.RegisteredClaims
}

// JWTService handles JWT operations
type JWTService struct {
	secretKey     []byte
	tokenDuration time.Duration
	issuer        string
}

// TokenBlacklist manages invalidated tokens
type TokenBlacklist struct {
	blacklist map[string]time.Time
	mu        sync.RWMutex
}

var (
	blacklist     *TokenBlacklist
	blacklistOnce sync.Once
)

// GetTokenBlacklist returns the singleton instance of TokenBlacklist
func GetTokenBlacklist() *TokenBlacklist {
	blacklistOnce.Do(func() {
		blacklist = &TokenBlacklist{
			blacklist: make(map[string]time.Time),
		}
	})
	return blacklist
}

// AddToBlacklist adds a token to the blacklist with its expiry time
func (tb *TokenBlacklist) AddToBlacklist(tokenString string, expiryTime time.Time) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.blacklist[tokenString] = expiryTime
	tb.cleanup() // Cleanup expired tokens
}

// IsBlacklisted checks if a token is blacklisted
func (tb *TokenBlacklist) IsBlacklisted(tokenString string) bool {
	tb.mu.RLock()
	defer tb.mu.RUnlock()
	_, exists := tb.blacklist[tokenString]
	return exists
}

// cleanup removes expired tokens from the blacklist
func (tb *TokenBlacklist) cleanup() {
	now := time.Now()
	for token, expiry := range tb.blacklist {
		if now.After(expiry) {
			delete(tb.blacklist, token)
		}
	}
}

// NewJWTService creates a new JWT service
func NewJWTService(config *config.Config) *JWTService {
	return &JWTService{
		secretKey:     []byte(config.Auth.JWTSecret),
		tokenDuration: time.Duration(config.Auth.JWTExpiryHours) * time.Hour,
		issuer:        config.Auth.JWTIssuer,
	}
}

// GenerateToken generates a new JWT token for a user
func GenerateToken(userID uuid.UUID, email string, roles []string, orgID uuid.UUID, permissions []string, secret string, expiryHours int) (string, error) {
	claims := Claims{
		UserID:      userID,
		Email:       email,
		Roles:       roles,
		OrgID:       orgID,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expiryHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// RefreshToken refreshes a JWT token
func (s *JWTService) RefreshToken(tokenString string) (string, error) {
	claims, err := ValidateToken(tokenString, string(s.secretKey))
	if err != nil {
		return "", err
	}

	// Check if token is about to expire
	now := time.Now()
	expiry := claims.ExpiresAt.Time
	threshold := expiry.Add(-6 * time.Hour) // Refresh if less than 6 hours left

	if now.Before(threshold) {
		return tokenString, nil // Token still valid for more than threshold
	}

	// Generate new token with same claims but new expiry
	return GenerateToken(
		claims.UserID,
		claims.Email,
		claims.Roles,
		claims.OrgID,
		claims.Permissions,
		string(s.secretKey),
		int(s.tokenDuration.Hours()),
	)
}

// GenerateTemporaryToken creates a temporary token for MFA verification
func GenerateTemporaryToken(userID uuid.UUID, email string, secret string, expiryHours int) (string, error) {
	// Create token
	token := jwt.New(jwt.SigningMethodHS256)

	// Set claims
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID.String()
	claims["email"] = email
	claims["temp"] = true // Mark as temporary token
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(expiryHours)).Unix()
	claims["iat"] = time.Now().Unix()
	claims["nbf"] = time.Now().Unix()

	// Generate signed token
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error generating token: %w", err)
	}

	return tokenString, nil
}

// ValidateTemporaryToken validates a temporary token and returns claims
func ValidateTemporaryToken(tokenString string, secret string) (jwt.MapClaims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	// Validate token
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("could not get token claims")
	}

	// Ensure it's a temporary token
	if temp, ok := claims["temp"].(bool); !ok || !temp {
		return nil, errors.New("not a temporary token")
	}

	return claims, nil
}
