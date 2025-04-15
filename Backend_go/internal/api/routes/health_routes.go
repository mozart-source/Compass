package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// HealthResponse represents the health check response structure
type HealthResponse struct {
	Status    string    `json:"status" example:"healthy"`
	Timestamp time.Time `json:"timestamp" example:"2025-04-17T02:00:00Z"`
}

// SetupHealthRoutes registers health check endpoints
func SetupHealthRoutes(router *gin.Engine) {
	// @Summary Health check endpoint
	// @Description Get the current health status of the API
	// @Tags health
	// @Produce json
	// @Success 200 {object} HealthResponse
	// @Router /health [get]
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{
			Status:    "healthy",
			Timestamp: time.Now().UTC(),
		})
	})

	// @Summary Readiness check endpoint
	// @Description Get the current readiness status of the API
	// @Tags health
	// @Produce json
	// @Success 200 {object} HealthResponse
	// @Router /health/ready [get]
	router.GET("/health/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, HealthResponse{
			Status:    "ready",
			Timestamp: time.Now().UTC(),
		})
	})
}
