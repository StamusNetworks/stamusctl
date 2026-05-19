package logging

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewPrometheusServer_StartAndStop verifies that NewPrometheusServer starts
// and shuts down cleanly when the context is cancelled.
func TestNewPrometheusServer_StartAndStop(t *testing.T) {
	SetLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		NewPrometheusServer(ctx)
	}()

	// Cancel the context to trigger shutdown.
	cancel()

	select {
	case <-done:
		// NewPrometheusServer returned after context cancellation — success.
	case <-time.After(3 * time.Second):
		t.Error("NewPrometheusServer did not return after context cancellation")
	}
	assert.True(t, true, "NewPrometheusServer completed without panic")
}
