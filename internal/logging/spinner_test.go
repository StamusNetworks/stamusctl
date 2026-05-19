package logging

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSpinner_ReturnsNonNil(t *testing.T) {
	s := NewSpinner("prefix-", "final\n")
	assert.NotNil(t, s)
	// Immediately stop so the background goroutine doesn't leak.
	s.Stop()
}

func TestSpinnerStop_NonNil(t *testing.T) {
	s := NewSpinner("", "")
	// Should not panic.
	SpinnerStop(s)
}

func TestSpinnerStop_Nil(t *testing.T) {
	// SpinnerStop must be a no-op for nil.
	SpinnerStop(nil)
}

func TestNewLogger_NoFileLogger(t *testing.T) {
	logger := NewLogger(false)
	assert.NotNil(t, logger)
}

func TestNewLogger_EnvTypePrd(t *testing.T) {
	original := envType
	envType = "prd"
	defer func() { envType = original }()

	logger := NewLogger(false)
	assert.NotNil(t, logger)
}

// TestCreateExporter_EmptyURL verifies that createExporter returns a non-nil
// exporter when called with an empty collector URL (stdout exporter path).
func TestCreateExporter_EmptyURL(t *testing.T) {
	// Ensure the global logger is initialised so Sugar.Fatal doesn't nil-deref.
	SetLogger()

	exporter := createExporter("")
	assert.NotNil(t, exporter)
	// Shutdown should be callable without error.
	_ = exporter.Shutdown(nil) //nolint:staticcheck // nil ctx OK for test
}

// TestInitTracer_EmptyCollector covers the branch where collectorURL is empty
// (stdout exporter) and returns the shutdown func.
func TestInitTracer_EmptyCollector(t *testing.T) {
	SetLogger()

	shutdown := InitTracer("", "test-service")
	assert.NotNil(t, shutdown)
}
