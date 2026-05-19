package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"stamus-ctl/pkg"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupHealthRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	NewHealth(r)
	return r
}

func TestHealthHandler_ReturnsOK(t *testing.T) {
	router := setupHealthRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response pkg.HealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "ok", response.Status)
	assert.Equal(t, "daemon is running", response.Message)
}

func TestHealthHandler_JSONContentType(t *testing.T) {
	router := setupHealthRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	assert.Contains(t, contentType, "application/json")
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	router := setupHealthRouter()

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(method, "/health", nil)
			router.ServeHTTP(w, req)

			// Gin returns 404 for unregistered method/path combinations
			assert.Equal(t, http.StatusNotFound, w.Code)
		})
	}
}

func TestNewHealth_RegistersRoutes(t *testing.T) {
	router := gin.New()
	NewHealth(router)

	// Verify /health route is registered
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

// TestReadinessHandler_ReturnsJSON verifies the readiness endpoint responds with JSON.
func TestReadinessHandler_ReturnsJSON(t *testing.T) {
	router := setupHealthRouter()

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/ready", nil)
	require.NoError(t, err)
	router.ServeHTTP(recorder, req)

	ct := recorder.Header().Get("Content-Type")
	assert.Contains(t, ct, "application/json")
}

// TestReadinessHandler_ResponseShape verifies the response has the expected fields.
func TestReadinessHandler_ResponseShape(t *testing.T) {
	router := setupHealthRouter()

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/ready", nil)
	require.NoError(t, err)
	router.ServeHTTP(recorder, req)

	var resp pkg.ReadinessResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.True(t, resp.Status == "ready" || resp.Status == "not_ready",
		"unexpected status: %q", resp.Status)
	assert.NotEmpty(t, resp.Message)
	assert.NotNil(t, resp.Checks)

	for _, key := range []string{"docker_daemon", "configuration", "resources"} {
		_, ok := resp.Checks[key]
		assert.True(t, ok, "checks map must contain %q key", key)
	}
}

// TestReadinessHandler_StatusCodeConsistent verifies HTTP status matches readiness status.
func TestReadinessHandler_StatusCodeConsistent(t *testing.T) {
	router := setupHealthRouter()

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "/ready", nil)
	require.NoError(t, err)
	router.ServeHTTP(recorder, req)

	var resp pkg.ReadinessResponse
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &resp))

	if resp.Status == "ready" {
		assert.Equal(t, http.StatusOK, recorder.Code)
	} else {
		assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	}
}

// TestCheckDockerConnectivity_ReturnsBoolean verifies no panic on Docker check.
func TestCheckDockerConnectivity_ReturnsBoolean(t *testing.T) {
	assert.NotPanics(t, func() {
		_ = checkDockerConnectivity()
	})
}
