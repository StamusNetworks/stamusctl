package shutdown

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewManager(t *testing.T) {
	m := NewManager(nil)
	if m == nil {
		t.Fatal("NewManager returned nil")
	}

	if m.IsShuttingDown() {
		t.Error("New manager should not be shutting down")
	}

	if m.timeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, m.timeout)
	}
}

func TestManagerSetTimeout(t *testing.T) {
	m := NewManager(nil)
	newTimeout := 10 * time.Second
	m.SetTimeout(newTimeout)

	if m.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, m.timeout)
	}
}

func TestManagerContext(t *testing.T) {
	m := NewManager(nil)
	ctx := m.Context()

	if ctx == nil {
		t.Fatal("Context should not be nil")
	}

	select {
	case <-ctx.Done():
		t.Error("Context should not be done before shutdown")
	default:
		// Expected
	}
}

func TestManagerRegister(t *testing.T) {
	m := NewManager(nil)
	called := false

	m.Register(Handler{
		Name:     "test-handler",
		Priority: PriorityFirst,
		Fn: func(ctx context.Context) error {
			called = true
			return nil
		},
	})

	m.Shutdown()

	if !called {
		t.Error("Handler should have been called during shutdown")
	}
}

func TestManagerHandlerPriorityOrder(t *testing.T) {
	m := NewManager(nil)
	var order []int
	var mu sync.Mutex

	// Register in reverse order to verify sorting
	m.Register(Handler{
		Name:     "last",
		Priority: PriorityLast,
		Fn: func(ctx context.Context) error {
			mu.Lock()
			order = append(order, 3)
			mu.Unlock()
			return nil
		},
	})

	m.Register(Handler{
		Name:     "first",
		Priority: PriorityFirst,
		Fn: func(ctx context.Context) error {
			mu.Lock()
			order = append(order, 1)
			mu.Unlock()
			return nil
		},
	})

	m.Register(Handler{
		Name:     "middle",
		Priority: PriorityConnections,
		Fn: func(ctx context.Context) error {
			mu.Lock()
			order = append(order, 2)
			mu.Unlock()
			return nil
		},
	})

	m.Shutdown()

	if len(order) != 3 {
		t.Fatalf("Expected 3 handlers to be called, got %d", len(order))
	}

	if order[0] != 1 || order[1] != 2 || order[2] != 3 {
		t.Errorf("Handlers called in wrong order: %v", order)
	}
}

func TestManagerShutdownIdempotent(t *testing.T) {
	m := NewManager(nil)
	var callCount atomic.Int32

	m.Register(Handler{
		Name:     "counter",
		Priority: PriorityFirst,
		Fn: func(ctx context.Context) error {
			callCount.Add(1)
			return nil
		},
	})

	// Call shutdown multiple times
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			m.Shutdown()
		}()
	}
	wg.Wait()

	if callCount.Load() != 1 {
		t.Errorf("Handler should be called exactly once, got %d calls", callCount.Load())
	}
}

func TestManagerTriggerShutdown(t *testing.T) {
	m := NewManager(nil)

	m.Register(Handler{
		Name:     "test",
		Priority: PriorityFirst,
		Fn: func(ctx context.Context) error {
			return nil
		},
	})

	m.TriggerShutdown(ExitError)

	<-m.Done()

	if !m.IsShuttingDown() {
		t.Error("Manager should be shutting down after TriggerShutdown")
	}
}

func TestManagerContextCancelledOnShutdown(t *testing.T) {
	m := NewManager(nil)
	ctx := m.Context()

	// Start shutdown in background
	go func() {
		time.Sleep(50 * time.Millisecond)
		m.Shutdown()
	}()

	select {
	case <-ctx.Done():
		// Expected
	case <-time.After(1 * time.Second):
		t.Error("Context should be cancelled on shutdown")
	}
}

func TestManagerRegisterDuringShutdown(t *testing.T) {
	logger := zap.NewNop()
	m := NewManager(logger)

	// Initiate shutdown
	m.Shutdown()
	<-m.Done()

	// Try to register after shutdown - should be rejected
	called := false
	m.Register(Handler{
		Name:     "late-handler",
		Priority: PriorityFirst,
		Fn: func(ctx context.Context) error {
			called = true
			return nil
		},
	})

	if called {
		t.Error("Handler registered during shutdown should not be called")
	}
}

// Operation Tracker Tests

func TestOperationTrackerStart(t *testing.T) {
	tracker := NewOperationTracker(nil, 5*time.Second)

	if tracker.ActiveCount() != 0 {
		t.Error("Initial active count should be 0")
	}

	ctx, done := tracker.Start(context.Background(), "test-op")

	if ctx == nil {
		t.Fatal("Start should return a context")
	}

	if tracker.ActiveCount() != 1 {
		t.Errorf("Active count should be 1, got %d", tracker.ActiveCount())
	}

	done()

	if tracker.ActiveCount() != 0 {
		t.Errorf("Active count should be 0 after done, got %d", tracker.ActiveCount())
	}
}

func TestOperationTrackerMultipleOperations(t *testing.T) {
	tracker := NewOperationTracker(nil, 5*time.Second)

	var dones []func()
	for i := 0; i < 5; i++ {
		_, done := tracker.Start(context.Background(), "op-"+string(rune('A'+i)))
		dones = append(dones, done)
	}

	if tracker.ActiveCount() != 5 {
		t.Errorf("Active count should be 5, got %d", tracker.ActiveCount())
	}

	for _, done := range dones {
		done()
	}

	if tracker.ActiveCount() != 0 {
		t.Errorf("Active count should be 0, got %d", tracker.ActiveCount())
	}
}

func TestOperationTrackerWaitForCompletion(t *testing.T) {
	tracker := NewOperationTracker(nil, 5*time.Second)

	// Start an operation that completes after 50ms
	_, done := tracker.Start(context.Background(), "test-op")

	go func() {
		time.Sleep(50 * time.Millisecond)
		done()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := tracker.WaitForCompletion(ctx)
	if err != nil {
		t.Errorf("WaitForCompletion failed: %v", err)
	}
}

func TestOperationTrackerWaitTimeout(t *testing.T) {
	tracker := NewOperationTracker(nil, 5*time.Second)

	// Start an operation that never completes
	_, _ = tracker.Start(context.Background(), "stuck-op")

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := tracker.WaitForCompletion(ctx)
	if err == nil {
		t.Error("WaitForCompletion should timeout")
	}
}

func TestOperationTrackerCancelAll(t *testing.T) {
	tracker := NewOperationTracker(nil, 5*time.Second)

	ctx1, done1 := tracker.Start(context.Background(), "op1")
	ctx2, done2 := tracker.Start(context.Background(), "op2")
	defer done1()
	defer done2()

	tracker.CancelAll()

	select {
	case <-ctx1.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context 1 should be cancelled")
	}

	select {
	case <-ctx2.Done():
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("Context 2 should be cancelled")
	}
}

func TestOperationTrackerShutdownHandler(t *testing.T) {
	tracker := NewOperationTracker(nil, 5*time.Second)

	handler := tracker.ShutdownHandler()

	if handler.Name != "operation-tracker" {
		t.Errorf("Expected handler name 'operation-tracker', got '%s'", handler.Name)
	}

	if handler.Priority != PriorityInFlight {
		t.Errorf("Expected priority %d, got %d", PriorityInFlight, handler.Priority)
	}

	// Test handler execution with no operations
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := handler.Fn(ctx)
	if err != nil {
		t.Errorf("Handler should complete successfully with no operations: %v", err)
	}
}

// Handler Priority Tests

func TestHandlerListSorting(t *testing.T) {
	handlers := handlerList{
		{Name: "last", Priority: PriorityLast},
		{Name: "first", Priority: PriorityFirst},
		{Name: "inflight", Priority: PriorityInFlight},
		{Name: "connections", Priority: PriorityConnections},
		{Name: "telemetry", Priority: PriorityTelemetry},
	}

	expected := []string{"first", "inflight", "connections", "telemetry", "last"}

	// Sort is implemented via sort.Interface
	for i := 0; i < len(handlers)-1; i++ {
		for j := i + 1; j < len(handlers); j++ {
			if handlers.Less(j, i) {
				handlers.Swap(i, j)
			}
		}
	}

	for i, h := range handlers {
		if h.Name != expected[i] {
			t.Errorf("Expected handler %s at position %d, got %s", expected[i], i, h.Name)
		}
	}
}

// Exit Codes Tests

func TestExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"Success", ExitSuccess, 0},
		{"Error", ExitError, 1},
		{"Misuse", ExitMisuse, 2},
		{"SIGINT", ExitSIGINT, 130},
		{"SIGTERM", ExitSIGTERM, 143},
		{"Timeout", ExitTimeout, 124},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("Expected %s to be %d, got %d", tt.name, tt.expected, tt.code)
			}
		})
	}
}

// Integration Tests

func TestManagerWithTracker(t *testing.T) {
	m := NewManager(nil)
	tracker := NewOperationTracker(nil, 5*time.Second)

	// Register tracker's handler
	m.Register(tracker.ShutdownHandler())

	// Start an operation
	_, done := tracker.Start(m.Context(), "test-op")

	// Complete operation in background
	go func() {
		time.Sleep(50 * time.Millisecond)
		done()
	}()

	// Trigger shutdown
	m.TriggerShutdown(ExitSuccess)

	// Wait for shutdown to complete
	select {
	case <-m.Done():
		// Expected
	case <-time.After(2 * time.Second):
		t.Error("Shutdown should complete within 2 seconds")
	}

	if tracker.ActiveCount() != 0 {
		t.Errorf("All operations should be complete, got %d active", tracker.ActiveCount())
	}
}

func TestListenForSignals_SIGTERM(t *testing.T) {
	m := NewManager(nil)

	m.Register(Handler{
		Name:     "noop",
		Priority: PriorityFirst,
		Fn: func(ctx context.Context) error {
			return nil
		},
	})

	exitCh := make(chan int, 1)
	go func() {
		code := m.ListenForSignals()
		exitCh <- code
	}()

	// Give the goroutine time to start listening.
	time.Sleep(20 * time.Millisecond)

	// Send SIGTERM to ourselves.
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("could not find current process: %v", err)
	}
	if err := p.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("failed to send SIGTERM: %v", err)
	}

	select {
	case code := <-exitCh:
		if code != ExitSIGTERM {
			t.Errorf("expected exit code %d (SIGTERM), got %d", ExitSIGTERM, code)
		}
	case <-time.After(3 * time.Second):
		t.Error("ListenForSignals did not return after SIGTERM")
	}
}

func TestListenForSignals_SIGINT(t *testing.T) {
	m := NewManager(nil)

	exitCh := make(chan int, 1)
	go func() {
		code := m.ListenForSignals()
		exitCh <- code
	}()

	time.Sleep(20 * time.Millisecond)

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatalf("could not find current process: %v", err)
	}
	if err := p.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("failed to send SIGINT: %v", err)
	}

	select {
	case code := <-exitCh:
		if code != ExitSIGINT {
			t.Errorf("expected exit code %d (SIGINT), got %d", ExitSIGINT, code)
		}
	case <-time.After(3 * time.Second):
		t.Error("ListenForSignals did not return after SIGINT")
	}
}

// TestManagerShutdown_TimeoutExceeded verifies the ctx.Err() != nil path in Shutdown().
// When the shutdown timeout is very short and a handler sleeps longer than the
// timeout, the second handler is skipped (timeout exceeded path).
func TestManagerShutdown_TimeoutExceeded(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	m := NewManager(logger)
	// Use a 10ms timeout so it expires while the slow handler is running.
	m.SetTimeout(10 * time.Millisecond)

	var slowCalled, fastCalled bool

	// Register a slow handler (priority=first, runs first) that sleeps > timeout.
	m.Register(Handler{
		Name:     "slow-handler",
		Priority: PriorityFirst,
		Fn: func(ctx context.Context) error {
			slowCalled = true
			// Sleep longer than the 10ms timeout so ctx expires before next handler.
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	})

	// Register a fast handler (priority=last, runs second) that should be skipped.
	m.Register(Handler{
		Name:     "fast-handler",
		Priority: PriorityLast,
		Fn: func(ctx context.Context) error {
			fastCalled = true
			return nil
		},
	})

	m.Shutdown()

	if !slowCalled {
		t.Error("slow handler should have been called")
	}
	// The fast handler should be skipped because the timeout context expired.
	if fastCalled {
		t.Error("fast handler should have been skipped due to timeout")
	}
}
