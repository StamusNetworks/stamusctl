package shutdown

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	globalManager *Manager
	globalTracker *OperationTracker
	once          sync.Once
)

// DefaultOperationTimeout is the default timeout for waiting on in-flight operations.
const DefaultOperationTimeout = 25 * time.Second

// Init initializes the global shutdown manager and operation tracker.
// This should be called early in application startup.
// If called multiple times, only the first call has effect.
func Init(logger *zap.Logger) {
	once.Do(func() {
		globalManager = NewManager(logger)
		globalTracker = NewOperationTracker(logger, DefaultOperationTimeout)

		// Register the operation tracker's shutdown handler
		globalManager.Register(globalTracker.ShutdownHandler())
	})
}

// GetManager returns the global shutdown manager.
// Panics if Init has not been called.
func GetManager() *Manager {
	if globalManager == nil {
		panic("shutdown: Init must be called before GetManager")
	}
	return globalManager
}

// GetTracker returns the global operation tracker.
// Panics if Init has not been called.
func GetTracker() *OperationTracker {
	if globalTracker == nil {
		panic("shutdown: Init must be called before GetTracker")
	}
	return globalTracker
}

// Context returns the global shutdown context.
// This context is cancelled when shutdown begins.
func Context() context.Context {
	return GetManager().Context()
}

// Register is a convenience function to register a handler with the global manager.
func Register(h Handler) {
	GetManager().Register(h)
}

// IsShuttingDown returns true if shutdown has been initiated.
func IsShuttingDown() bool {
	if globalManager == nil {
		return false
	}
	return globalManager.IsShuttingDown()
}
