package shutdown

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// OperationTracker tracks in-flight operations that should complete before shutdown.
type OperationTracker struct {
	mu         sync.RWMutex
	operations map[string]context.CancelFunc
	wg         sync.WaitGroup
	count      atomic.Int64
	logger     *zap.Logger
	timeout    time.Duration
}

// NewOperationTracker creates a new operation tracker.
func NewOperationTracker(logger *zap.Logger, timeout time.Duration) *OperationTracker {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &OperationTracker{
		operations: make(map[string]context.CancelFunc),
		logger:     logger,
		timeout:    timeout,
	}
}

// Start registers a new operation and returns a context that will be cancelled on shutdown.
// The operation ID is used for logging and tracking.
// The returned done function MUST be called when the operation completes.
func (t *OperationTracker) Start(ctx context.Context, operationID string) (context.Context, func()) {
	opCtx, cancel := context.WithCancel(ctx)

	t.mu.Lock()
	t.operations[operationID] = cancel
	t.mu.Unlock()

	t.wg.Add(1)
	count := t.count.Add(1)
	t.logger.Debug("Operation started",
		zap.String("operation", operationID),
		zap.Int64("active_count", count))

	var once sync.Once
	done := func() {
		once.Do(func() {
			t.mu.Lock()
			delete(t.operations, operationID)
			t.mu.Unlock()

			t.wg.Done()
			newCount := t.count.Add(-1)
			t.logger.Debug("Operation completed",
				zap.String("operation", operationID),
				zap.Int64("active_count", newCount))
		})
	}

	return opCtx, done
}

// ActiveCount returns the number of currently active operations.
func (t *OperationTracker) ActiveCount() int64 {
	return t.count.Load()
}

// WaitForCompletion waits for all tracked operations to complete.
// It respects the context deadline and returns an error if the timeout is exceeded.
func (t *OperationTracker) WaitForCompletion(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		t.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		t.logger.Info("All in-flight operations completed")
		return nil
	case <-ctx.Done():
		activeCount := t.count.Load()
		t.logger.Warn("Timeout waiting for operations to complete",
			zap.Int64("remaining_operations", activeCount))
		return ctx.Err()
	}
}

// CancelAll cancels all tracked operations.
// This does not wait for operations to complete; use WaitForCompletion for that.
func (t *OperationTracker) CancelAll() {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for id, cancel := range t.operations {
		t.logger.Debug("Cancelling operation", zap.String("operation", id))
		cancel()
	}
}

// ShutdownHandler returns a Handler that waits for operations to complete.
func (t *OperationTracker) ShutdownHandler() Handler {
	return Handler{
		Name:     "operation-tracker",
		Priority: PriorityInFlight,
		Fn: func(ctx context.Context) error {
			// Create a timeout context for waiting
			waitCtx, cancel := context.WithTimeout(ctx, t.timeout)
			defer cancel()

			return t.WaitForCompletion(waitCtx)
		},
	}
}
