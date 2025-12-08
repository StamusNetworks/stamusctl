package docker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryConfig(t *testing.T) {
	config := GetRetryConfig()
	if config == nil {
		t.Fatal("GetRetryConfig returned nil")
	}

	if config.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts to be 3, got %d", config.MaxAttempts)
	}

	if config.InitialBackoff != time.Second {
		t.Errorf("Expected InitialBackoff to be 1s, got %v", config.InitialBackoff)
	}

	if config.MaxBackoff != 30*time.Second {
		t.Errorf("Expected MaxBackoff to be 30s, got %v", config.MaxBackoff)
	}

	if config.BackoffMultiplier != 2.0 {
		t.Errorf("Expected BackoffMultiplier to be 2.0, got %f", config.BackoffMultiplier)
	}
}

func TestSetRetryConfig(t *testing.T) {
	originalConfig := GetRetryConfig()

	customConfig := &RetryConfig{
		MaxAttempts:       5,
		InitialBackoff:    2 * time.Second,
		MaxBackoff:        60 * time.Second,
		BackoffMultiplier: 3.0,
	}

	SetRetryConfig(customConfig)

	config := GetRetryConfig()
	if config.MaxAttempts != 5 {
		t.Errorf("Expected MaxAttempts to be 5, got %d", config.MaxAttempts)
	}

	// Restore original config
	SetRetryConfig(originalConfig)
}

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	// Initially closed
	if cb.GetState() != "closed" {
		t.Errorf("Expected initial state to be closed, got %s", cb.GetState())
	}

	// Can attempt when closed
	if err := cb.CanAttempt(); err != nil {
		t.Errorf("Expected to be able to attempt when closed: %v", err)
	}

	// Record failures
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.GetState() != "closed" {
		t.Errorf("Expected state to be closed after 2 failures, got %s", cb.GetState())
	}

	// Third failure should open circuit
	cb.RecordFailure()
	if cb.GetState() != "open" {
		t.Errorf("Expected state to be open after 3 failures, got %s", cb.GetState())
	}

	// Should not be able to attempt when open
	if err := cb.CanAttempt(); err == nil {
		t.Error("Expected error when attempting with open circuit")
	}

	// Wait for timeout to transition to half-open
	time.Sleep(1100 * time.Millisecond)
	if err := cb.CanAttempt(); err != nil {
		t.Errorf("Expected to transition to half-open after timeout: %v", err)
	}
	if cb.GetState() != "half-open" {
		t.Errorf("Expected state to be half-open after timeout, got %s", cb.GetState())
	}

	// Success should close circuit
	cb.RecordSuccess()
	if cb.GetState() != "closed" {
		t.Errorf("Expected state to be closed after success in half-open, got %s", cb.GetState())
	}
}

func TestCircuitBreakerReset(t *testing.T) {
	cb := NewCircuitBreaker(2, 1*time.Second)

	// Open the circuit
	cb.RecordFailure()
	cb.RecordFailure()
	if cb.GetState() != "open" {
		t.Errorf("Expected state to be open, got %s", cb.GetState())
	}

	// Reset should close circuit
	cb.Reset()
	if cb.GetState() != "closed" {
		t.Errorf("Expected state to be closed after reset, got %s", cb.GetState())
	}
}

func TestIsTransientError(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"connection refused", errors.New("connection refused"), true},
		{"connection reset", errors.New("connection reset by peer"), true},
		{"timeout", errors.New("i/o timeout"), true},
		{"EOF", errors.New("EOF"), true},
		{"no such host", errors.New("no such host"), true},
		{"network unreachable", errors.New("network is unreachable"), true},
		{"TLS timeout", errors.New("TLS handshake timeout"), true},
		{"broken pipe", errors.New("broken pipe"), true},
		{"non-transient error", errors.New("invalid argument"), false},
		{"image not found", errors.New("image not found"), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isTransientError(tc.err)
			if result != tc.expected {
				t.Errorf("Expected isTransientError(%v) to be %v, got %v", tc.err, tc.expected, result)
			}
		})
	}
}

func TestWithRetrySuccess(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	cb := NewCircuitBreaker(5, 1*time.Second)
	attempts := 0

	operation := func() error {
		attempts++
		return nil
	}

	err := WithRetry(ctx, config, cb, operation, "test-success")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}

	if cb.GetState() != "closed" {
		t.Errorf("Expected circuit breaker to be closed, got %s", cb.GetState())
	}
}

func TestWithRetryTransientError(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	cb := NewCircuitBreaker(5, 1*time.Second)
	attempts := 0

	operation := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("connection refused")
		}
		return nil
	}

	err := WithRetry(ctx, config, cb, operation, "test-transient")
	if err != nil {
		t.Errorf("Expected no error after retries, got %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestWithRetryNonTransientError(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	cb := NewCircuitBreaker(5, 1*time.Second)
	attempts := 0

	operation := func() error {
		attempts++
		return errors.New("invalid argument")
	}

	err := WithRetry(ctx, config, cb, operation, "test-non-transient")
	if err == nil {
		t.Error("Expected error for non-transient failure")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt (no retry for non-transient), got %d", attempts)
	}
}

func TestWithRetryMaxAttemptsExceeded(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	cb := NewCircuitBreaker(5, 1*time.Second)
	attempts := 0

	operation := func() error {
		attempts++
		return errors.New("connection timeout")
	}

	err := WithRetry(ctx, config, cb, operation, "test-max-attempts")
	if err == nil {
		t.Error("Expected error after max attempts")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestWithRetryContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &RetryConfig{
		MaxAttempts:       5,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        1 * time.Second,
		BackoffMultiplier: 2.0,
	}

	cb := NewCircuitBreaker(5, 1*time.Second)
	attempts := 0

	operation := func() error {
		attempts++
		return errors.New("connection timeout")
	}

	// Cancel context after first attempt
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := WithRetry(ctx, config, cb, operation, "test-context-cancel")
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}

	// Should have attempted at least once, but not all 5 attempts
	if attempts < 1 || attempts >= 5 {
		t.Errorf("Expected 1-4 attempts due to context cancellation, got %d", attempts)
	}
}

func TestWithRetryExponentialBackoff(t *testing.T) {
	ctx := context.Background()
	config := &RetryConfig{
		MaxAttempts:       3,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        500 * time.Millisecond,
		BackoffMultiplier: 2.0,
	}

	cb := NewCircuitBreaker(5, 1*time.Second)
	attempts := 0
	startTime := time.Now()

	operation := func() error {
		attempts++
		if attempts < 3 {
			return errors.New("timeout")
		}
		return nil
	}

	err := WithRetry(ctx, config, cb, operation, "test-backoff")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	elapsed := time.Since(startTime)
	// Should take at least 100ms + 200ms = 300ms for 2 retries
	if elapsed < 300*time.Millisecond {
		t.Errorf("Expected at least 300ms elapsed, got %v", elapsed)
	}
}

func TestWithRetrySimple(t *testing.T) {
	attempts := 0
	operation := func() error {
		attempts++
		if attempts < 2 {
			return errors.New("connection refused")
		}
		return nil
	}

	err := WithRetrySimple(operation, "test-simple")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}
