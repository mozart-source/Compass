package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/logger"
	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/security/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var log = logger.NewLogger()

const (
	bearerSchema = "Bearer "
)

// RateLimiterConfig holds configuration for rate limiting
type RateLimiterConfig struct {
	Window      time.Duration
	MaxAttempts int64
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Error("Missing authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, bearerSchema) {
			log.Error("Invalid authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := authHeader[len(bearerSchema):]

		// Check if token is blacklisted
		if auth.GetTokenBlacklist().IsBlacklisted(tokenString) {
			log.Error("Token is blacklisted")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token has been invalidated"})
			c.Abort()
			return
		}

		// First validate the JWT token
		claims, err := auth.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			log.Error("Token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Check for service-to-service call indicator
		serviceToService := c.GetHeader("X-Service-Call") == "true" ||
			c.GetHeader("X-Internal-Service") != "" ||
			isServiceToServiceUA(c.GetHeader("User-Agent"))

		// For service-to-service calls, skip session validation if JWT is valid
		if serviceToService {
			log.Info("Service-to-service call detected, skipping session validation",
				zap.String("user_agent", c.GetHeader("User-Agent")),
				zap.String("user_id", claims.UserID.String()))

			// Store claims and token in context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("roles", claims.Roles)
			c.Set("org_id", claims.OrgID)
			c.Set("permissions", claims.Permissions)
			c.Set("token", tokenString)
			c.Set("is_service_call", true)

			c.Next()
			return
		}

		// For regular user requests, validate session
		session, exists := auth.GetSessionStore().GetSession(tokenString)
		if !exists {
			log.Error("Invalid or expired session")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired session"})
			c.Abort()
			return
		}

		// Update session activity
		auth.GetSessionStore().UpdateSessionActivity(tokenString)

		// Verify session user matches token user
		if session.UserID != claims.UserID {
			log.Error("Session user mismatch")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			c.Abort()
			return
		}

		// Store claims, token, and session in context
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("roles", claims.Roles)
		c.Set("org_id", claims.OrgID)
		c.Set("permissions", claims.Permissions)
		c.Set("token", tokenString)
		c.Set("session", session)

		c.Next()
	}
}

// isServiceToServiceUA checks if the User-Agent indicates a service-to-service call
func isServiceToServiceUA(userAgent string) bool {
	serviceUAs := []string{
		"aiohttp",
		"python-requests",
		"go-http-client",
		"curl",
		"HTTPie",
		"service",
		"backend",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, ua := range serviceUAs {
		if strings.Contains(userAgentLower, ua) {
			return true
		}
	}
	return false
}

// RateLimitMiddleware creates a middleware for rate limiting using Redis
func RateLimitMiddleware(limiter auth.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		path := c.Request.URL.Path
		key := fmt.Sprintf("%s:%s", ip, path)

		allowed, remaining, resetTime, err := limiter.Allow(c.Request.Context(), key)
		if err != nil {
			log.Error("Rate limiter error", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			c.Abort()
			return
		}

		if !allowed {
			c.Header("X-RateLimit-Reset", resetTime.String())
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":    "rate limit exceeded",
				"reset_in": time.Until(resetTime).String(),
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", resetTime.String())

		c.Next()
	}
}

// GetUserID retrieves the authenticated user's ID from the context
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, false
	}
	return userID.(uuid.UUID), true
}

// RequireRoles middleware checks if user has all required roles
func RequireRoles(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRolesVal, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		userRoles, ok := userRolesVal.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid roles format in token"})
			c.Abort()
			return
		}

		// Create a map for efficient lookup of user's roles
		userRolesMap := make(map[string]struct{})
		for _, role := range userRoles {
			userRolesMap[role] = struct{}{}
		}

		// Check if the user has all the required roles
		for _, requiredRole := range roles {
			if _, found := userRolesMap[requiredRole]; !found {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// RequirePermissions middleware checks if user has all required permissions
func RequirePermissions(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPermissionsVal, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			c.Abort()
			return
		}

		userPermissions, ok := userPermissionsVal.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid permissions format in token"})
			c.Abort()
			return
		}

		// Create a map for efficient lookup of user's permissions
		userPermissionsMap := make(map[string]struct{})
		for _, perm := range userPermissions {
			userPermissionsMap[perm] = struct{}{}
		}

		// Check if the user has all the required permissions
		for _, requiredPerm := range permissions {
			if _, found := userPermissionsMap[requiredPerm]; !found {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// JWTAuthMiddleware creates a middleware for JWT authentication
func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return NewAuthMiddleware(jwtSecret)
}
