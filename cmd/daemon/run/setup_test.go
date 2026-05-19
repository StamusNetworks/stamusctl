package run

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"stamus-ctl/internal/logging"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	logging.SetLogger()
	m.Run()
}

// TestSetupRouter_BasicRoutes verifies that SetupRouter registers expected routes.
// Not marked parallel because SetupRouter writes to redisClient (global).
func TestSetupRouter_BasicRoutes(t *testing.T) {
	logging.InitTracer("", "test-service")
	logger := func(_ string) {}

	router := SetupRouter(context.Background(), logger)
	require.NotNil(t, router)

	routes := router.Routes()
	routeSet := make(map[string]bool)
	for _, route := range routes {
		routeSet[route.Method+" "+route.Path] = true
	}

	assert.True(t, routeSet["GET /health"], "/health GET must be registered")
	assert.True(t, routeSet["GET /ready"], "/ready GET must be registered")
}

// TestSetupRouter_PingHealthEndpoint verifies /health is reachable (no token configured).
func TestSetupRouter_PingHealthEndpoint(t *testing.T) {
	logging.InitTracer("", "test-service")
	logger := func(_ string) {}

	router := SetupRouter(context.Background(), logger)

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil)
	require.NoError(t, err)
	router.ServeHTTP(recorder, req)

	// /health is always accessible without auth
	assert.Equal(t, http.StatusOK, recorder.Code)
}

// TestRateLimiter_ReturnsHandler verifies RateLimiter returns a non-nil middleware handler.
func TestRateLimiter_ReturnsHandler(t *testing.T) {
	// Not parallel — RateLimiter writes to redisClient global.
	handler := RateLimiter()
	assert.NotNil(t, handler)
}

// TestRateLimiter_AllowsRequests verifies the rate limiter middleware actually processes requests.
func TestRateLimiter_AllowsRequests(t *testing.T) {
	// Not parallel — RateLimiter writes to redisClient global.
	router := gin.New()
	router.Use(RateLimiter())
	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	require.NoError(t, err)
	router.ServeHTTP(recorder, req)

	// First request should succeed (200) or be rate-limited (429)
	assert.True(t, recorder.Code == http.StatusOK || recorder.Code == http.StatusTooManyRequests,
		"expected 200 or 429, got %d", recorder.Code)
}

// TestErrorHandler_WritesResponse verifies the error handler formats the response correctly.
func TestErrorHandler_WritesResponse(t *testing.T) {
	t.Parallel()

	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		info := ratelimit.Info{ //nolint:exhaustruct // only ResetTime matters for this handler
			ResetTime: time.Now().Add(5 * time.Second),
		}
		errorHandler(c, info)
	})

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "/test", nil)
	require.NoError(t, err)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Too many requests")
}

// TestGetLogger_ReturnsFunc verifies that getLogger returns a callable function.
func TestGetLogger_ReturnsFunc(t *testing.T) {
	t.Parallel()

	logging.InitTracer("", "test-service")
	_, span := logging.Tracer.Start(context.Background(), "test")

	defer span.End()

	logFn := getLogger(span)
	require.NotNil(t, logFn)

	// Calling the returned function must not panic.
	assert.NotPanics(t, func() {
		logFn("test message")
	})
}

// TestSetupLogging_ReturnsSpan verifies that setupLogging returns a non-nil span.
func TestSetupLogging_ReturnsSpan(t *testing.T) {
	t.Parallel()

	logging.InitTracer("", "test-service")

	span := setupLogging(context.Background())
	assert.NotNil(t, span)
}

// TestSetupRouter_WithTokenPath exercises the tokenpath != "" branches in SetupRouter.
// It uses a real temp file as the token path to avoid WatchForToken failing.
func TestSetupRouter_WithTokenPath(t *testing.T) {
	logging.InitTracer("", "test-tokenpath")

	// Create a temp token file.
	tmpFile, err := os.CreateTemp("", "test-token-*.txt")
	require.NoError(t, err)
	tmpFile.WriteString("test-token-content")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Set tokenpath so SetupRouter takes the authenticated branches.
	viper.Set("tokenpath", tmpFile.Name())
	defer viper.Set("tokenpath", "")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := func(_ string) {}
	router := SetupRouter(ctx, logger)
	require.NotNil(t, router)

	// Cancel to stop WatchForToken goroutine.
	cancel()
}
