package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/internal/infrastructure/cache"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CacheMiddleware struct {
	cache  *cache.RedisClient
	prefix string
	ttl    time.Duration
}

func NewCacheMiddleware(cache *cache.RedisClient, prefix string, ttl time.Duration) *CacheMiddleware {
	return &CacheMiddleware{
		cache:  cache,
		prefix: prefix,
		ttl:    ttl,
	}
}

// responseBuffer is a custom ResponseWriter that stores the response
type responseBuffer struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func newResponseBuffer(original gin.ResponseWriter) *responseBuffer {
	return &responseBuffer{
		ResponseWriter: original,
		body:           bytes.NewBufferString(""),
	}
}

func (r *responseBuffer) Write(b []byte) (int, error) {
	r.ResponseWriter.Write(b)
	return r.body.Write(b)
}

func (r *responseBuffer) WriteString(s string) (int, error) {
	r.ResponseWriter.WriteString(s)
	return r.body.WriteString(s)
}

func (r *responseBuffer) WriteHeader(code int) {
	r.ResponseWriter.WriteHeader(code)
}

// CacheResponse caches the response of an endpoint
func (m *CacheMiddleware) CacheResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Generate cache key
		key := m.generateCacheKey(c)

		// Try to get from cache
		if cached, err := m.cache.Get(c, key); err == nil {
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				c.JSON(http.StatusOK, response)
				c.Abort()
				return
			}
		}

		// Store original response writer
		writer := c.Writer
		// Create a copy buffer with the original writer
		buff := newResponseBuffer(writer)
		c.Writer = buff

		// Process request
		c.Next()

		// If response was successful, cache it
		if c.Writer.Status() == http.StatusOK {
			responseData := buff.body.String()
			if err := m.cache.Set(c, key, responseData, m.ttl); err != nil {
				log.Error("Failed to cache response", zap.Error(err))
			}
		}

		// Original writer already has the response due to our WriteHeader and Write implementations
		c.Writer = writer
	}
}

// CacheInvalidate invalidates cache entries based on patterns
func (m *CacheMiddleware) CacheInvalidate(patterns ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			for _, pattern := range patterns {
				key := fmt.Sprintf("%s:%s", m.prefix, pattern)
				if err := m.cache.ClearByPattern(c, key); err != nil {
					log.Error("Failed to invalidate cache", zap.Error(err), zap.String("pattern", pattern))
				}
			}
		}
	}
}

func (m *CacheMiddleware) generateCacheKey(c *gin.Context) string {
	// Get user ID from context for user-specific caching
	userID, _ := GetUserID(c)

	// Build key from method, path, query params, and user ID
	parts := []string{m.prefix}

	// Extract resource type and ID from path
	pathParts := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")
	if len(pathParts) >= 2 {
		resourceType := pathParts[1] // e.g., "tasks"
		parts = append(parts, resourceType)

		// If this is a specific resource (has ID)
		if len(pathParts) >= 3 {
			resourceID := pathParts[2]
			if _, err := uuid.Parse(resourceID); err == nil {
				parts = append(parts, "id", resourceID)
			} else {
				parts = append(parts, "list")
			}
		} else {
			parts = append(parts, "list")
		}
	}

	// Add sorted query parameters
	if len(c.Request.URL.RawQuery) > 0 {
		parts = append(parts, c.Request.URL.RawQuery)
	}

	// Add user ID for user-specific caching
	if userID != uuid.Nil {
		parts = append(parts, userID.String())
	}

	return strings.Join(parts, ":")
}

// CachePageWithTTL caches the response of an endpoint with a custom TTL
func (m *CacheMiddleware) CachePageWithTTL(keyPrefix string, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "GET" {
			c.Next()
			return
		}

		// Generate cache key with the provided prefix
		userID, _ := GetUserID(c)
		key := m.generateCacheKeyWithPrefix(keyPrefix, c, userID.String())

		// Try to get from cache
		if cached, err := m.cache.Get(c, key); err == nil {
			var response map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				c.JSON(http.StatusOK, response)
				c.Abort()
				return
			}
		}

		// Store original response writer
		writer := c.Writer
		// Create a copy buffer with the original writer
		buff := newResponseBuffer(writer)
		c.Writer = buff

		// Process request
		c.Next()

		// If response was successful, cache it
		if c.Writer.Status() == http.StatusOK {
			responseData := buff.body.String()
			if err := m.cache.Set(c, key, responseData, ttl); err != nil {
				log.Error("Failed to cache response", zap.Error(err))
			}
		}

		// Original writer already has the response due to our WriteHeader and Write implementations
		c.Writer = writer
	}
}

func (m *CacheMiddleware) generateCacheKeyWithPrefix(prefix string, c *gin.Context, userID string) string {
	// Build key from prefix, method, path, query params, and user ID
	parts := []string{m.prefix, prefix}

	// Add resource information from the path
	pathParts := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")
	if len(pathParts) >= 2 {
		resourceType := pathParts[1] // e.g., "tasks"
		parts = append(parts, resourceType)

		// If this is a specific resource (has ID)
		if len(pathParts) >= 3 {
			resourceID := pathParts[2]
			if _, err := uuid.Parse(resourceID); err == nil {
				parts = append(parts, "id", resourceID)
			} else {
				parts = append(parts, "list")
			}
		} else {
			parts = append(parts, "list")
		}
	}

	// Add sorted query parameters
	if len(c.Request.URL.RawQuery) > 0 {
		parts = append(parts, c.Request.URL.RawQuery)
	}

	// Add user ID for user-specific caching
	if userID != "" {
		parts = append(parts, userID)
	}

	return strings.Join(parts, ":")
}
