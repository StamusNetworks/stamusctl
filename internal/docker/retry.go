package docker

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"stamus-ctl/internal/logging"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxAttempts       int           // Maximum number of retry attempts
	InitialBackoff    time.Duration // Initial backoff duration
	MaxBackoff        time.Duration // Maximum backoff duration
	BackoffMultiplier float64       // Multiplier for exponential backoff
}

// CircuitBreaker prevents retry storms by tracking failures and opening when threshold is reached
type CircuitBreaker struct {
	mu              sync.Mutex
	failureCount    int
	successCount    int
	lastFailureTime time.Time
	state           string // "closed", "open", "half-open"
	threshold       int    // Failure threshold before opening
	timeout         time.Duration
}

var (
	// DefaultRetryConfig provides sensible defaults for retry behavior
	defaultRetryConfig = &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
	}

	// Global circuit breaker for Docker operations
	dockerCircuitBreaker = NewCircuitBreaker(5, 30*time.Second)
)

// GetRetryConfig returns the current retry configuration
func GetRetryConfig() *RetryConfig {
	return defaultRetryConfig
}

// SetRetryConfig allows customizing retry behavior
func SetRetryConfig(config *RetryConfig) {
	if config != nil {
		defaultRetryConfig = config
	}
}

// NewCircuitBreaker creates a new circuit breaker with specified threshold and timeout
func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:     "closed",
		threshold: threshold,
		timeout:   timeout,
	}
}

// CanAttempt checks if a request can be attempted based on circuit breaker state
func (cb *CircuitBreaker) CanAttempt() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case "open":
		if time.Since(cb.lastFailureTime) > cb.timeout {
			cb.state = "half-open"
			logging.Sugar.Info("circuit breaker entering half-open state")
			return nil
		}
		return fmt.Errorf("circuit breaker is open, preventing retry storm")
	case "half-open", "closed":
		return nil
	}
	return nil
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.successCount++
	cb.failureCount = 0

	if cb.state == "half-open" {
		cb.state = "closed"
		logging.Sugar.Info("circuit breaker closed after successful attempt")
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.threshold {
		if cb.state != "open" {
			cb.state = "open"
			logging.Sugar.Warnw("circuit breaker opened", "failures", cb.failureCount)
		}
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = "closed"
	cb.failureCount = 0
	cb.successCount = 0
}

// isTransientError determines if an error is transient and should be retried
func isTransientError(err error) bool {
	if err == nil {
		return false
	}

	// Common transient error patterns for Docker and network operations
	transientPatterns := []string{
		"connection refused",
		"connection reset",
		"timeout",
		"temporary failure",
		"TLS handshake timeout",
		"no such host",
		"network is unreachable",
		"EOF",
		"broken pipe",
		"i/o timeout",
		"client is newer than server",
		"context deadline exceeded",
		"transport is closing",
		"use of closed network connection",
		"error reading from server",
		"received unexpected HTTP status",
	}

	errStr := strings.ToLower(err.Error())
	for _, pattern := range transientPatterns {
		if strings.Contains(errStr, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// WithRetry executes a function with retry logic and exponential backoff
func WithRetry(ctx context.Context, config *RetryConfig, cb *CircuitBreaker, operation func() error, operationName string) error {
	if config == nil {
		config = defaultRetryConfig
	}

	if cb == nil {
		cb = dockerCircuitBreaker
	}

	logger := logging.Sugar.With("operation", operationName)

	var lastErr error
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Check circuit breaker
		if err := cb.CanAttempt(); err != nil {
			logger.Warnw("circuit breaker prevented attempt", "error", err, "state", cb.GetState())
			return fmt.Errorf("circuit breaker open: %w", err)
		}

		// Execute operation
		err := operation()
		if err == nil {
			cb.RecordSuccess()
			if attempt > 0 {
				logger.Infow("operation succeeded after retry", "attempts", attempt+1)
			}
			return nil
		}

		lastErr = err

		// Check if error is transient
		if !isTransientError(err) {
			logger.Debugw("non-transient error, not retrying", "error", err)
			cb.RecordFailure()
			return err
		}

		// Last attempt, don't wait
		if attempt == config.MaxAttempts-1 {
			logger.Warnw("max retry attempts reached", "attempts", config.MaxAttempts, "error", err)
			cb.RecordFailure()
			return fmt.Errorf("max retries exceeded: %w", err)
		}

		// Calculate exponential backoff
		backoff := time.Duration(float64(config.InitialBackoff) *
			math.Pow(config.BackoffMultiplier, float64(attempt)))
		if backoff > config.MaxBackoff {
			backoff = config.MaxBackoff
		}

		logger.Debugw("retrying after backoff", "attempt", attempt+1, "backoff", backoff, "error", err)

		// Wait for backoff or context cancellation
		select {
		case <-time.After(backoff):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		}
	}

	cb.RecordFailure()
	return fmt.Errorf("operation failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// WithRetrySimple is a convenience wrapper that uses default config and circuit breaker
func WithRetrySimple(operation func() error, operationName string) error {
	return WithRetry(ctx, defaultRetryConfig, dockerCircuitBreaker, operation, operationName)
}
