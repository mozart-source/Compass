package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	// Header names
	RequestIDHeader    = "X-Request-ID"
	TraceIDHeader      = "X-Trace-ID"
	SpanIDHeader       = "X-Span-ID"
	ParentSpanIDHeader = "X-Parent-Span-ID"
)

// TraceContext holds tracing information
type TraceContext struct {
	RequestID    string
	TraceID      string
	SpanID       string
	ParentSpanID string
	StartTime    time.Time
}

// TracingMiddleware implements distributed tracing
type TracingMiddleware struct {
	log *logger.Logger
}

// NewTracingMiddleware creates a new tracing middleware
func NewTracingMiddleware() *TracingMiddleware {
	return &TracingMiddleware{
		log: logger.NewLogger(),
	}
}

// TraceRequest adds tracing information to requests
func (m *TracingMiddleware) TraceRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get or generate request ID
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = generateID()
		}

		// Get or generate trace ID
		traceID := c.GetHeader(TraceIDHeader)
		if traceID == "" {
			traceID = generateID()
		}

		// Get or generate span ID
		spanID := c.GetHeader(SpanIDHeader)
		if spanID == "" {
			spanID = generateID()
		}

		// Get parent span ID
		parentSpanID := c.GetHeader(ParentSpanIDHeader)

		// Create trace context
		traceCtx := &TraceContext{
			RequestID:    requestID,
			TraceID:      traceID,
			SpanID:       spanID,
			ParentSpanID: parentSpanID,
			StartTime:    time.Now(),
		}

		// Store trace context in gin context
		c.Set("trace_context", traceCtx)

		// Add trace headers to response
		c.Header(RequestIDHeader, requestID)
		c.Header(TraceIDHeader, traceID)
		c.Header(SpanIDHeader, spanID)
		if parentSpanID != "" {
			c.Header(ParentSpanIDHeader, parentSpanID)
		}

		// Log request start
		m.log.Info("Request started",
			zap.String("request_id", requestID),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
			zap.String("parent_span_id", parentSpanID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(traceCtx.StartTime)

		// Log request completion
		m.log.Info("Request completed",
			zap.String("request_id", requestID),
			zap.String("trace_id", traceID),
			zap.String("span_id", spanID),
			zap.String("parent_span_id", parentSpanID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
		)
	}
}

// GetTraceContext retrieves the trace context from the gin context
func GetTraceContext(c *gin.Context) *TraceContext {
	if traceCtx, exists := c.Get("trace_context"); exists {
		return traceCtx.(*TraceContext)
	}
	return nil
}

// generateID generates a random ID
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
