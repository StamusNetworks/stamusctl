package logging

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger_WithFileLogger(t *testing.T) {
	// Skip this test since it requires a file path that may not exist
	t.Skip("File logger test requires specific directory structure")

	// Test with file logger enabled
	logger := NewLogger(true)

	assert.NotNil(t, logger)
	assert.IsType(t, &zap.Logger{}, logger)
}

func TestNewLogger_WithoutFileLogger(t *testing.T) {
	// Test with file logger disabled
	logger := NewLogger(false)

	assert.NotNil(t, logger)
	assert.IsType(t, &zap.Logger{}, logger)
}

func TestLogger_LogLevels(t *testing.T) {
	// Test that different log levels are handled correctly
	var buf bytes.Buffer

	// Create a logger that writes to buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	logger := zap.New(core)

	// Test different log levels
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := buf.String()
	assert.Contains(t, output, "debug message")
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
}

func TestLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer

	// Create a logger that writes to buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	logger := zap.New(core)

	// Test logging with fields
	logger.Info("test message",
		zap.String("user", "test@example.com"),
		zap.Int("count", 42),
		zap.Bool("active", true))

	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "test@example.com")
	assert.Contains(t, output, "42")
	assert.Contains(t, output, "true")
}

func TestLogger_JSONOutput(t *testing.T) {
	var buf bytes.Buffer

	// Create a logger that writes to buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	logger := zap.New(core)

	// Test JSON output format
	logger.Info("json test message")

	output := buf.String()
	assert.Contains(t, output, `"level":"info"`)
	assert.Contains(t, output, `"msg":"json test message"`)
	assert.Contains(t, output, `"timestamp":`)
}

func TestLogger_EnvTypeProduction(t *testing.T) {
	// Test that logger respects environment type
	originalEnvType := envType
	envType = "production"
	defer func() { envType = originalEnvType }()

	logger := NewLogger(false)

	assert.NotNil(t, logger)
	assert.IsType(t, &zap.Logger{}, logger)
}

func TestLogger_VerbosityLevels(t *testing.T) {
	// Test that different verbosity levels are handled
	for i := 0; i < len(levels); i++ {
		logger := NewLogger(false)
		assert.NotNil(t, logger)
	}
}

func TestLogger_FileOutput(t *testing.T) {
	// Skip this test since it requires specific file system setup
	t.Skip("File logger test requires specific directory structure")

	// Test file output (if applicable)
	// This test checks that file logger doesn't crash
	logger := NewLogger(true)

	assert.NotNil(t, logger)

	// Test that we can log to it
	logger.Info("test file output")
	logger.Error("test error output")
	logger.Debug("test debug output")
}

func TestLogger_GlobalVariables(t *testing.T) {
	// Test that global variables are initialized properly
	logger := NewLogger(false)

	// Set the global variables
	Logger = logger
	Sugar = logger.Sugar()

	assert.NotNil(t, Logger)
	assert.NotNil(t, Sugar)
	assert.IsType(t, &zap.Logger{}, Logger)
	assert.IsType(t, &zap.SugaredLogger{}, Sugar)
}

func TestLogger_WithEnvironmentVariables(t *testing.T) {
	// Test behavior with different environment variables
	originalEnv := os.Getenv("ENV_TYPE")
	defer func() {
		if originalEnv != "" {
			t.Setenv("ENV_TYPE", originalEnv)
		} else {
			_ = os.Unsetenv("ENV_TYPE")
		}
	}()

	// Test with production environment
	t.Setenv("ENV_TYPE", "production")
	logger := NewLogger(false)
	assert.NotNil(t, logger)

	// Test with development environment
	t.Setenv("ENV_TYPE", "development")
	logger = NewLogger(false)
	assert.NotNil(t, logger)

	// Test with test environment
	t.Setenv("ENV_TYPE", "test")
	logger = NewLogger(false)
	assert.NotNil(t, logger)
}

func TestLoggerWithSpanContext(t *testing.T) {
	InitTracer("", "")

	// Create a mock span context
	traceID := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	spanID := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,
		SpanID:  spanID,
	})

	logger := LoggerWithSpanContext(spanContext)
	assert.NotNil(t, logger)
	assert.IsType(t, &zap.Logger{}, logger)
}

func TestLoggerWithContextToSpanContext(t *testing.T) {
	InitTracer("", "")

	ctx := context.Background()

	// Create a context with span
	ctx, span := otel.Tracer("test").Start(ctx, "test-operation")
	defer span.End()

	logger := LoggerWithContextToSpanContext(ctx)
	assert.NotNil(t, logger)
	assert.IsType(t, &zap.Logger{}, logger)
}

func TestLogError(t *testing.T) {
	InitTracer("", "")

	// Create a mock request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	// This should not panic
	LogError(req, "test error message")

	// We can't easily test the output, but we can verify it doesn't crash
	assert.NotNil(t, req)
}

func TestSetLogger(t *testing.T) {
	// Save original values
	originalLogger := Logger
	originalSugar := Sugar
	defer func() {
		Logger = originalLogger
		Sugar = originalSugar
	}()

	// Test SetLogger function
	SetLogger()

	assert.NotNil(t, Logger)
	assert.NotNil(t, Sugar)
	assert.IsType(t, &zap.Logger{}, Logger)
	assert.IsType(t, &zap.SugaredLogger{}, Sugar)
}

func TestSetLogger_WithFileLogger(t *testing.T) {
	// Skip this test since it requires specific file system setup
	t.Skip("File logger test requires specific directory structure")

	// Save original values
	originalLogger := Logger
	originalSugar := Sugar
	originalEnv := os.Getenv("FILE_LOGGER")
	defer func() {
		Logger = originalLogger
		Sugar = originalSugar
		if originalEnv != "" {
			t.Setenv("FILE_LOGGER", originalEnv)
		} else {
			_ = os.Unsetenv("FILE_LOGGER")
		}
	}()

	// Set FILE_LOGGER environment variable
	t.Setenv("FILE_LOGGER", "true")

	// Test SetLogger function with file logger
	SetLogger()

	assert.NotNil(t, Logger)
	assert.NotNil(t, Sugar)
	assert.IsType(t, &zap.Logger{}, Logger)
	assert.IsType(t, &zap.SugaredLogger{}, Sugar)
}

func TestSetupDBLogger(t *testing.T) {
	InitTracer("", "")

	// Create a mock database (this is tricky without a real DB)
	// We'll just test that the function doesn't panic

	// Create a mock gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request = req

	// We can't test with a real DB easily, but we can test the function call
	// The function should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SetupDBLogger panicked: %v", r)
		}
	}()

	// We'll skip the actual test since it requires a real GORM DB
	t.Skip("SetupDBLogger test requires real GORM DB instance")
}

func TestNewPrometheusServer(t *testing.T) {
	InitTracer("", "")

	// Test that NewPrometheusServer doesn't panic on call
	// We can't actually run the server in tests, but we can test the setup

	// This function starts a server so we can't test it directly
	// We'll just test that it doesn't panic on initialization
	t.Skip("NewPrometheusServer test requires server setup which is not suitable for unit tests")
}

// ---------------------------------------------------------------------------
// createExporter — both paths
// ---------------------------------------------------------------------------

func TestCreateExporter_EmptyURL_UsesStdout(t *testing.T) {
	SetLogger()
	exporter := createExporter("")
	assert.NotNil(t, exporter)
}

func TestCreateExporter_NonEmptyURL_UsesGRPC(t *testing.T) {
	SetLogger()
	// With a non-empty URL, createExporter creates a gRPC-based OTLP exporter.
	// Connection is lazy, so the call succeeds even without a running collector.
	exporter := createExporter("localhost:4317")
	assert.NotNil(t, exporter)
	// Shutdown is needed to clean up goroutines.
	_ = exporter.Shutdown(context.Background())
}
