package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

// ValidationMiddleware handles request validation
type ValidationMiddleware struct {
	validator *validator.Validate
	log       *logger.Logger
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	v := validator.New()

	// Register custom validators
	v.RegisterValidation("not_empty", validateNotEmpty)
	v.RegisterValidation("valid_uuid", validateUUID)

	return &ValidationMiddleware{
		validator: v,
		log:       logger.NewLogger(),
	}
}

// ValidateRequest validates the request body against the provided struct
func (m *ValidationMiddleware) ValidateRequest(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Make a new instance of the model to bind to
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		modelValue := reflect.New(modelType).Interface()

		// Read the request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
		}

		// Log the raw request for debugging
		m.log.Info("Request details",
			zap.String("path", c.Request.URL.Path),
			zap.String("content_type", c.GetHeader("Content-Type")),
			zap.Int64("content_length", c.Request.ContentLength),
			zap.String("body", string(bodyBytes)))

		// Restore the request body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// Try to bind JSON
		if err := json.Unmarshal(bodyBytes, modelValue); err != nil {
			m.log.Error("JSON unmarshal failed",
				zap.Error(err),
				zap.String("body", string(bodyBytes)))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Invalid JSON format: %v", err.Error()),
			})
			c.Abort()
			return
		}

		// Validate model
		if err := m.validator.Struct(modelValue); err != nil {
			// Format validation errors
			errors := make(map[string]string)
			for _, err := range err.(validator.ValidationErrors) {
				field := strings.ToLower(err.Field())
				errors[field] = formatValidationError(err)
			}

			m.log.Error("Validation failed",
				zap.Any("errors", errors),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation failed",
				"details": errors,
			})
			c.Abort()
			return
		}

		// Store validated model in context
		c.Set("validated_model", modelValue)
		c.Next()
	}
}

// ValidateQuery validates query parameters against the provided struct
func (m *ValidationMiddleware) ValidateQuery(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create new instance of model
		modelType := reflect.TypeOf(model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		modelValue := reflect.New(modelType).Interface()

		// Bind query parameters
		if err := c.ShouldBindQuery(modelValue); err != nil {
			m.log.Error("Failed to bind query parameters",
				zap.Error(err),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid query parameters",
			})
			c.Abort()
			return
		}

		// Validate model
		if err := m.validator.Struct(modelValue); err != nil {
			// Format validation errors
			errors := make(map[string]string)
			for _, err := range err.(validator.ValidationErrors) {
				field := strings.ToLower(err.Field())
				errors[field] = formatValidationError(err)
			}

			m.log.Error("Query validation failed",
				zap.Any("errors", errors),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "validation failed",
				"details": errors,
			})
			c.Abort()
			return
		}

		// Store validated model in context
		c.Set("validated_query", modelValue)
		c.Next()
	}
}

// Custom validators
func validateNotEmpty(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return len(strings.TrimSpace(value)) > 0
}

func validateUUID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	// Basic UUID format validation
	return len(value) == 36 && strings.Count(value, "-") == 4
}

// Helper function to format validation errors
func formatValidationError(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "this field is required"
	case "email":
		return "invalid email format"
	case "min":
		return "value is too short"
	case "max":
		return "value is too long"
	case "not_empty":
		return "this field cannot be empty"
	case "valid_uuid":
		return "invalid UUID format"
	default:
		return "invalid value"
	}
}
