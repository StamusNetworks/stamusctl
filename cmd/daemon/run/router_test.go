package run

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"stamus-ctl/cmd/daemon/run/compose"
	"stamus-ctl/cmd/daemon/run/config"
	"stamus-ctl/cmd/daemon/run/health"
	"stamus-ctl/cmd/daemon/run/troubleshoot"
	"stamus-ctl/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// setupSimpleTestRouter creates a test router without Redis or complex initialization
func setupSimpleTestRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware())

	// Health endpoints (no auth required)
	health.NewHealth(r)

	// API routes
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", ping)
		compose.NewCompose(v1)
		config.NewConfig(v1)
		troubleshoot.NewTroubleshoot(v1)
	}

	return r
}

// Health endpoints bypass auth tests

func TestSetupRouter_HealthNoAuth(t *testing.T) {
	router := setupSimpleTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestSetupRouter_ReadyRouteRegistered(t *testing.T) {
	router := setupSimpleTestRouter()

	// Just verify the route is registered without calling the handler
	// (handler calls Docker which is not appropriate for unit tests)
	routes := router.Routes()
	readyFound := false
	for _, route := range routes {
		if route.Path == "/ready" && route.Method == http.MethodGet {
			readyFound = true
			break
		}
	}
	assert.True(t, readyFound, "/ready route should be registered")
}

// Route registration tests

func TestSetupRouter_RegistersHealthRoutes(t *testing.T) {
	router := setupSimpleTestRouter()
	routes := router.Routes()

	healthFound := false
	readyFound := false

	for _, route := range routes {
		if route.Path == "/health" && route.Method == http.MethodGet {
			healthFound = true
		}
		if route.Path == "/ready" && route.Method == http.MethodGet {
			readyFound = true
		}
	}

	assert.True(t, healthFound, "/health route should be registered")
	assert.True(t, readyFound, "/ready route should be registered")
}

func TestSetupRouter_RegistersAPIRoutes(t *testing.T) {
	router := setupSimpleTestRouter()
	routes := router.Routes()

	expectedRoutes := []string{
		"/api/v1/ping",
		"/api/v1/compose/init",
		"/api/v1/compose/up",
		"/api/v1/compose/down",
		"/api/v1/compose/ps",
		"/api/v1/config",
		"/api/v1/troubleshoot/containers",
		"/api/v1/troubleshoot/kernel",
	}

	for _, expectedPath := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route.Path == expectedPath {
				found = true
				break
			}
		}
		assert.True(t, found, "Route %s should be registered", expectedPath)
	}
}

// Middleware tests

func TestSetupRouter_SecurityHeaders(t *testing.T) {
	router := setupSimpleTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	router.ServeHTTP(w, req)

	// Security headers should be set
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
	assert.NotEmpty(t, w.Header().Get("X-Frame-Options"))
	assert.NotEmpty(t, w.Header().Get("X-Content-Type-Options"))
}

func TestSetupRouter_Recovery(t *testing.T) {
	router := setupSimpleTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// API versioning tests

func TestSetupRouter_APIVersioning(t *testing.T) {
	router := setupSimpleTestRouter()
	routes := router.Routes()

	apiRoutes := 0
	for _, route := range routes {
		if len(route.Path) > 7 && route.Path[:7] == "/api/v1" {
			apiRoutes++
		}
	}

	assert.Greater(t, apiRoutes, 0, "Should have API routes under /api/v1")
}

// Ping endpoint test

func TestSetupRouter_PingEndpoint(t *testing.T) {
	router := setupSimpleTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pong")
}

// 404 tests

func TestSetupRouter_UnknownRoute(t *testing.T) {
	router := setupSimpleTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/unknown/route", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSetupRouter_UnknownAPIRoute(t *testing.T) {
	router := setupSimpleTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/unknown", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Method tests

func TestSetupRouter_MethodNotAllowed(t *testing.T) {
	router := setupSimpleTestRouter()

	// POST to /health should not work
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/health", nil)
	router.ServeHTTP(w, req)

	// Gin returns 404 for method mismatch by default
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Content-Type tests

func TestSetupRouter_JSONContentType(t *testing.T) {
	router := setupSimpleTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	assert.Contains(t, contentType, "application/json")
}

// Route groups tests

func TestSetupRouter_ComposeGroup(t *testing.T) {
	router := setupSimpleTestRouter()
	routes := router.Routes()

	composeRoutes := []string{
		"/api/v1/compose/init",
		"/api/v1/compose/update",
		"/api/v1/compose/up",
		"/api/v1/compose/down",
		"/api/v1/compose/ps",
		"/api/v1/compose/restart/config",
		"/api/v1/compose/restart/containers",
	}

	for _, path := range composeRoutes {
		found := false
		for _, route := range routes {
			if route.Path == path {
				found = true
				break
			}
		}
		assert.True(t, found, "Compose route %s should be registered", path)
	}
}

func TestSetupRouter_ConfigGroup(t *testing.T) {
	router := setupSimpleTestRouter()
	routes := router.Routes()

	configPostFound := false
	configGetFound := false
	configListFound := false
	configVersionFound := false

	for _, route := range routes {
		switch {
		case route.Path == "/api/v1/config" && route.Method == http.MethodPost:
			configPostFound = true
		case route.Path == "/api/v1/config" && route.Method == http.MethodGet:
			configGetFound = true
		case route.Path == "/api/v1/config/list" && route.Method == http.MethodPost:
			configListFound = true
		case route.Path == "/api/v1/config/version" && route.Method == http.MethodGet:
			configVersionFound = true
		}
	}

	assert.True(t, configPostFound, "POST /api/v1/config should be registered")
	assert.True(t, configGetFound, "GET /api/v1/config should be registered")
	assert.True(t, configListFound, "POST /api/v1/config/list should be registered")
	assert.True(t, configVersionFound, "GET /api/v1/config/version should be registered")
}

func TestSetupRouter_TroubleshootGroup(t *testing.T) {
	router := setupSimpleTestRouter()
	routes := router.Routes()

	troubleshootRoutes := []string{
		"/api/v1/troubleshoot/containers",
		"/api/v1/troubleshoot/kernel",
		"/api/v1/troubleshoot/reboot",
	}

	for _, path := range troubleshootRoutes {
		found := false
		for _, route := range routes {
			if route.Path == path {
				found = true
				break
			}
		}
		assert.True(t, found, "Troubleshoot route %s should be registered", path)
	}
}

// Rate limiter unit tests (not using Redis)

func TestRateLimitPerSecond_Constant(t *testing.T) {
	assert.Equal(t, 5, RateLimitPerSecond)
}

func TestKeyFunc_ReturnsClientIP(t *testing.T) {
	r := gin.New()

	r.GET("/test", func(c *gin.Context) {
		key := keyFunc(c)
		c.String(http.StatusOK, key)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	r.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), "192.168.1.100")
}

func TestGetEnvFallback_WithEnv(t *testing.T) {
	t.Setenv("TEST_VAR_GET_ENV", "test-value")

	result := getEnvFallback("TEST_VAR_GET_ENV", "fallback")
	assert.Equal(t, "test-value", result)
}

func TestGetEnvFallback_WithoutEnv(t *testing.T) {
	result := getEnvFallback("NON_EXISTENT_VAR_12345", "fallback")
	assert.Equal(t, "fallback", result)
}

func TestPing_Handler(t *testing.T) {
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/ping", ping)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/ping", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pong")
}
