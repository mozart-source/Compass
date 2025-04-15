package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/ahmedelhadi17776/Compass/Backend_go/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreakerConfig holds configuration for the circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold    int           // Number of failures before opening circuit
	SuccessThreshold    int           // Number of successes before closing circuit
	Timeout             time.Duration // Time to wait before attempting to close circuit
	HalfOpenMaxRequests int           // Maximum number of requests in half-open state
}

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	config    CircuitBreakerConfig
	state     CircuitState
	failures  int
	successes int
	lastError time.Time
	mutex     sync.RWMutex
	log       *logger.Logger
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
		log:    logger.NewLogger(),
	}
}

// CircuitBreakerMiddleware creates a middleware that implements the circuit breaker pattern
func (cb *CircuitBreaker) CircuitBreakerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cb.mutex.RLock()
		state := cb.state
		cb.mutex.RUnlock()

		switch state {
		case StateOpen:
			// Check if timeout has elapsed
			cb.mutex.Lock()
			if time.Since(cb.lastError) > cb.config.Timeout {
				cb.state = StateHalfOpen
				cb.failures = 0
				cb.successes = 0
				cb.mutex.Unlock()
			} else {
				cb.mutex.Unlock()
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "service temporarily unavailable",
				})
				c.Abort()
				return
			}

		case StateHalfOpen:
			cb.mutex.Lock()
			if cb.successes >= cb.config.SuccessThreshold {
				cb.state = StateClosed
				cb.failures = 0
				cb.successes = 0
			} else if cb.failures >= cb.config.FailureThreshold {
				cb.state = StateOpen
				cb.lastError = time.Now()
				cb.mutex.Unlock()
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "service temporarily unavailable",
				})
				c.Abort()
				return
			}
			cb.mutex.Unlock()
		}

		// Process request
		c.Next()

		// Update circuit state based on response
		cb.mutex.Lock()
		defer cb.mutex.Unlock()

		if c.Writer.Status() >= 500 {
			cb.failures++
			if cb.state == StateClosed && cb.failures >= cb.config.FailureThreshold {
				cb.state = StateOpen
				cb.lastError = time.Now()
				cb.log.Error("Circuit breaker opened",
					zap.String("path", c.Request.URL.Path),
					zap.Int("failures", cb.failures))
			}
		} else {
			cb.successes++
			if cb.state == StateHalfOpen && cb.successes >= cb.config.SuccessThreshold {
				cb.state = StateClosed
				cb.failures = 0
				cb.successes = 0
				cb.log.Info("Circuit breaker closed",
					zap.String("path", c.Request.URL.Path))
			}
		}
	}
}
