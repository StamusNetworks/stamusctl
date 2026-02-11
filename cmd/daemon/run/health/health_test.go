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
