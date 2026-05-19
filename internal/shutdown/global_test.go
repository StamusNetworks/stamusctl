package shutdown

import (
	"context"
	"sync"
	"testing"

	"go.uber.org/zap"
)

// resetGlobals resets the package-level globals between tests so Init can be
// called fresh. This is safe because all tests in this package are sequential.
func resetGlobals() {
	globalManager = nil
	globalTracker = nil
	once = sync.Once{}
}

// ---------------------------------------------------------------------------
// TestInit_Idempotency
// ---------------------------------------------------------------------------

func TestInit_Idempotency(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	logger := zap.NewNop()

	Init(logger)
	m1 := globalManager
	tracker1 := globalTracker

	// Second call must be a no-op (sync.Once).
	Init(logger)
	m2 := globalManager
	tracker2 := globalTracker

	if m1 != m2 {
		t.Error("Init called twice produced different managers — expected the same pointer")
	}
	if tracker1 != tracker2 {
		t.Error("Init called twice produced different trackers — expected the same pointer")
	}
}

func TestInit_NilLogger(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	Init(nil)

	if globalManager == nil {
		t.Fatal("globalManager should be non-nil after Init")
	}
	if globalTracker == nil {
		t.Fatal("globalTracker should be non-nil after Init")
	}
}

// ---------------------------------------------------------------------------
// TestIsShuttingDown
// ---------------------------------------------------------------------------

func TestIsShuttingDown_BeforeInit(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	if IsShuttingDown() {
		t.Error("IsShuttingDown should return false when globalManager is nil")
	}
}

func TestIsShuttingDown_AfterInit_NotShuttingDown(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	Init(zap.NewNop())

	if IsShuttingDown() {
		t.Error("IsShuttingDown should return false immediately after Init")
	}
}

func TestIsShuttingDown_AfterShutdown(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	Init(zap.NewNop())
	globalManager.Shutdown()
	<-globalManager.Done()

	if !IsShuttingDown() {
		t.Error("IsShuttingDown should return true after Shutdown")
	}
}

// ---------------------------------------------------------------------------
// TestGetManager_AfterInit
// ---------------------------------------------------------------------------

func TestGetManager_AfterInit(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	Init(zap.NewNop())

	m := GetManager()
	if m == nil {
		t.Fatal("GetManager returned nil after Init")
	}
}

func TestGetManager_PanicsBeforeInit(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	defer func() {
		if r := recover(); r == nil {
			t.Error("GetManager should panic when called before Init")
		}
	}()

	GetManager()
}

// ---------------------------------------------------------------------------
// TestGetTracker_AfterInit
// ---------------------------------------------------------------------------

func TestGetTracker_AfterInit(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	Init(zap.NewNop())

	tracker := GetTracker()
	if tracker == nil {
		t.Fatal("GetTracker returned nil after Init")
	}
}

func TestGetTracker_PanicsBeforeInit(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	defer func() {
		if r := recover(); r == nil {
			t.Error("GetTracker should panic when called before Init")
		}
	}()

	GetTracker()
}

// ---------------------------------------------------------------------------
// TestContext_AfterInit
// ---------------------------------------------------------------------------

func TestContext_AfterInit(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	Init(zap.NewNop())

	ctx := Context()
	if ctx == nil {
		t.Fatal("Context() returned nil after Init")
	}

	select {
	case <-ctx.Done():
		t.Error("Context should not be done before shutdown begins")
	default:
		// expected
	}
}

func TestContext_CancelledAfterShutdown(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	Init(zap.NewNop())
	ctx := Context()

	globalManager.Shutdown()
	<-globalManager.Done()

	select {
	case <-ctx.Done():
		// expected
	default:
		t.Error("Context should be cancelled after shutdown")
	}
}

// ---------------------------------------------------------------------------
// TestRegister_ConvenienceFunction
// ---------------------------------------------------------------------------

func TestRegister_ConvenienceFunction(t *testing.T) {
	resetGlobals()
	defer resetGlobals()

	Init(zap.NewNop())

	called := false
	Register(Handler{
		Name:     "test-convenience",
		Priority: PriorityFirst,
		Fn: func(_ context.Context) error {
			called = true
			return nil
		},
	})

	globalManager.Shutdown()
	<-globalManager.Done()

	if !called {
		t.Error("Handler registered via Register convenience function was not called")
	}
}
