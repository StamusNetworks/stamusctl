package shutdown

import (
	"context"
	"os"
	"os/signal"
	"sort"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// State represents the current shutdown state.
type State int32

const (
	// StateRunning indicates normal operation.
	StateRunning State = iota
	// StateStopping indicates shutdown is in progress.
	StateStopping
	// StateStopped indicates shutdown is complete.
	StateStopped
)

// DefaultTimeout is the default maximum time allowed for graceful shutdown.
const DefaultTimeout = 30 * time.Second

// Manager coordinates graceful shutdown of all registered handlers.
type Manager struct {
	state    atomic.Int32
	handlers handlerList
	mu       sync.RWMutex
	timeout  time.Duration
	logger   *zap.Logger

	// ctx is cancelled when shutdown begins.
	ctx    context.Context
	cancel context.CancelFunc

	// done is closed when shutdown completes.
	done chan struct{}

	// exitCode stores the exit code to use.
	exitCode atomic.Int32

	// forceQuitCount tracks consecutive interrupt signals for force quit.
	forceQuitCount atomic.Int32
	lastSignalTime atomic.Int64
}

// NewManager creates a new shutdown manager with the given logger.
// If logger is nil, a no-op logger is used.
func NewManager(logger *zap.Logger) *Manager {
	if logger == nil {
		logger = zap.NewNop()
	}

	ctx, cancel := context.WithCancel(context.Background())

	m := &Manager{
		timeout:  DefaultTimeout,
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
		done:     make(chan struct{}),
		handlers: make(handlerList, 0),
	}
	m.state.Store(int32(StateRunning))
	m.exitCode.Store(int32(ExitSuccess))

	return m
}

// SetTimeout configures the maximum time allowed for graceful shutdown.
func (m *Manager) SetTimeout(d time.Duration) {
	m.mu.Lock()
	m.timeout = d
	m.mu.Unlock()
}

// Context returns a context that is cancelled when shutdown begins.
// Use this context for long-running operations that should be interrupted on shutdown.
func (m *Manager) Context() context.Context {
	return m.ctx
}

// Done returns a channel that is closed when shutdown completes.
func (m *Manager) Done() <-chan struct{} {
	return m.done
}

// IsShuttingDown returns true if shutdown has been initiated.
func (m *Manager) IsShuttingDown() bool {
	return State(m.state.Load()) != StateRunning
}

// Register adds a cleanup handler to be called during shutdown.
// Handlers are executed in priority order (lower priority numbers first).
// This method is safe to call concurrently.
func (m *Manager) Register(h Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IsShuttingDown() {
		m.logger.Warn("Cannot register handler during shutdown", zap.String("handler", h.Name))
		return
	}

	m.handlers = append(m.handlers, h)
	sort.Sort(m.handlers)
	m.logger.Debug("Registered shutdown handler", zap.String("handler", h.Name), zap.Int("priority", int(h.Priority)))
}

// ListenForSignals starts listening for OS signals (SIGTERM, SIGINT) and triggers shutdown.
// This method blocks until shutdown is complete.
// Double SIGINT within 2 seconds forces immediate exit.
func (m *Manager) ListenForSignals() int {
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case sig := <-sigChan:
			now := time.Now().UnixNano()
			lastTime := m.lastSignalTime.Swap(now)

			// Check for double-signal force quit (within 2 seconds)
			if sig == syscall.SIGINT {
				if now-lastTime < int64(2*time.Second) {
					count := m.forceQuitCount.Add(1)
					if count >= 2 && m.IsShuttingDown() {
						m.logger.Warn("Received second interrupt signal, forcing immediate exit")
						return ExitSIGINT
					}
				} else {
					m.forceQuitCount.Store(0)
				}
			}

			// Determine exit code based on signal
			var exitCode int
			switch sig {
			case syscall.SIGTERM:
				m.logger.Info("Received SIGTERM, initiating graceful shutdown")
				exitCode = ExitSIGTERM
			case syscall.SIGINT:
				m.logger.Info("Received SIGINT, initiating graceful shutdown")
				exitCode = ExitSIGINT
			}

			// Only initiate shutdown once
			if !m.IsShuttingDown() {
				m.exitCode.Store(int32(exitCode))
				go m.Shutdown()
			}

		case <-m.done:
			return int(m.exitCode.Load())
		}
	}
}

// Shutdown initiates graceful shutdown, executing all registered handlers.
// This method can be called directly for programmatic shutdown.
// It is safe to call multiple times; only the first call has effect.
func (m *Manager) Shutdown() {
	// Ensure we only shutdown once
	if !m.state.CompareAndSwap(int32(StateRunning), int32(StateStopping)) {
		return
	}

	m.mu.RLock()
	timeout := m.timeout
	m.mu.RUnlock()

	m.logger.Info("Starting graceful shutdown", zap.Duration("timeout", timeout))

	// Signal all operations to stop
	m.cancel()

	// Create timeout context for cleanup
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Execute handlers in priority order
	m.mu.RLock()
	handlers := make(handlerList, len(m.handlers))
	copy(handlers, m.handlers)
	m.mu.RUnlock()

	start := time.Now()
	for _, h := range handlers {
		if ctx.Err() != nil {
			m.logger.Warn("Shutdown timeout exceeded, skipping remaining handlers",
				zap.String("skipped_handler", h.Name))
			m.exitCode.Store(int32(ExitTimeout))
			break
		}

		m.logger.Debug("Executing shutdown handler", zap.String("handler", h.Name))
		handlerStart := time.Now()

		if err := h.Fn(ctx); err != nil {
			m.logger.Error("Shutdown handler failed",
				zap.String("handler", h.Name),
				zap.Error(err),
				zap.Duration("duration", time.Since(handlerStart)))
		} else {
			m.logger.Debug("Shutdown handler completed",
				zap.String("handler", h.Name),
				zap.Duration("duration", time.Since(handlerStart)))
		}
	}

	m.state.Store(int32(StateStopped))
	m.logger.Info("Graceful shutdown complete", zap.Duration("total_duration", time.Since(start)))
	close(m.done)
}

// TriggerShutdown initiates shutdown with a specific exit code.
// Useful for programmatic shutdown scenarios.
func (m *Manager) TriggerShutdown(exitCode int) {
	m.exitCode.Store(int32(exitCode))
	m.Shutdown()
}
