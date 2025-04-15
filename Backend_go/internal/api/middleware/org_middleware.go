package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OrganizationMiddleware extracts organization ID from header and sets it in context
func OrganizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := c.GetHeader("X-Organization-ID")
		if orgID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "X-Organization-ID header is required"})
			c.Abort()
			return
		}

		// Validate UUID format
		_, err := uuid.Parse(orgID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization ID format"})
			c.Abort()
			return
		}

		c.Set("organization_id", orgID)
		c.Next()
	}
}
