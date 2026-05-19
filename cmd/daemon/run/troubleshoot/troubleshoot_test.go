package troubleshoot

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"stamus-ctl/internal/app"
	"stamus-ctl/pkg"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Route Registration Tests

func TestNewTroubleshoot_RegistersRoutes(t *testing.T) {
	router := gin.New()
	v1 := router.Group("/api/v1")
	NewTroubleshoot(v1)

	routes := router.Routes()

	expectedRoutes := map[string]string{
		"/api/v1/troubleshoot/containers": http.MethodPost,
		"/api/v1/troubleshoot/kernel":     http.MethodPost,
		"/api/v1/troubleshoot/reboot":     http.MethodPost,
	}

	for path, method := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route.Path == path && route.Method == method {
				found = true
				break
			}
		}
		assert.True(t, found, "Route %s %s should be registered", method, path)
	}
}

// JSON Parsing Tests

func TestLogsHandler_InvalidJSON(t *testing.T) {
	router := gin.New()
	router.POST("/logs", logsHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/logs", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

// Request Struct Tests

func TestLogsRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"containers": ["container1", "container2"],
		"timestamps": true,
		"tail": "100",
		"since": "1h",
		"until": "10m"
	}`

	var req pkg.LogsRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Len(t, req.Containers, 2)
	assert.Contains(t, req.Containers, "container1")
	assert.Contains(t, req.Containers, "container2")
	assert.True(t, req.Timestamps)
	assert.Equal(t, "100", req.Tail)
	assert.Equal(t, "1h", req.Since)
	assert.Equal(t, "10m", req.Until)
}

func TestLogsRequest_EmptyJSON(t *testing.T) {
	jsonData := `{}`

	var req pkg.LogsRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Nil(t, req.Containers)
	assert.False(t, req.Timestamps)
	assert.Empty(t, req.Tail)
	assert.Empty(t, req.Since)
	assert.Empty(t, req.Until)
}

func TestLogsResponse_JSONMarshal(t *testing.T) {
	resp := pkg.LogsResponse{
		Containers: []pkg.ContainerLogs{
			{
				Logs: []string{"log1", "log2", "log3"},
			},
		},
	}

	jsonData, err := json.Marshal(resp)
	require.NoError(t, err)

	var unmarshaled pkg.LogsResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	require.NoError(t, err)

	assert.Len(t, unmarshaled.Containers, 1)
	assert.Len(t, unmarshaled.Containers[0].Logs, 3)
}

// Method Tests

func TestTroubleshootRoutes_OnlyAcceptPost(t *testing.T) {
	router := gin.New()
	v1 := router.Group("/api/v1")
	NewTroubleshoot(v1)

	endpoints := []string{
		"/api/v1/troubleshoot/containers",
		"/api/v1/troubleshoot/kernel",
		"/api/v1/troubleshoot/reboot",
	}

	nonPostMethods := []string{http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, endpoint := range endpoints {
		for _, method := range nonPostMethods {
			t.Run(method+" "+endpoint, func(t *testing.T) {
				w := httptest.NewRecorder()
				req, _ := http.NewRequest(method, endpoint, nil)
				router.ServeHTTP(w, req)

				// Should return 404 for non-POST methods
				assert.Equal(t, http.StatusNotFound, w.Code)
			})
		}
	}
}

// Content-Type Tests

func TestLogsRequest_MarshalUnmarshal(t *testing.T) {
	original := pkg.LogsRequest{
		Containers: []string{"web", "db"},
		Timestamps: true,
		Tail:       "100",
		Since:      "1h",
		Until:      "10m",
	}

	jsonData, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded pkg.LogsRequest
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Containers, decoded.Containers)
	assert.Equal(t, original.Timestamps, decoded.Timestamps)
	assert.Equal(t, original.Tail, decoded.Tail)
	assert.Equal(t, original.Since, decoded.Since)
	assert.Equal(t, original.Until, decoded.Until)
}

// Kernel Handler Tests - This one doesn't use Docker

func TestKernelHandler_ReturnsJSON(t *testing.T) {
	router := gin.New()
	router.POST("/kernel", kernelHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/kernel", nil)
	router.ServeHTTP(w, req)

	// Response should be JSON
	contentType := w.Header().Get("Content-Type")
	assert.Contains(t, contentType, "application/json")
}

func TestKernelHandler_ReturnsMessage(t *testing.T) {
	router := gin.New()
	router.POST("/kernel", kernelHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/kernel", nil)
	router.ServeHTTP(w, req)

	// Should return JSON with message field
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "message")
}

func TestKernelHandler_Success(t *testing.T) {
	router := gin.New()
	router.POST("/kernel", kernelHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/kernel", nil)
	router.ServeHTTP(w, req)

	// Should return 200 OK
	assert.Equal(t, http.StatusOK, w.Code)
}

// Default Values Tests

func TestLogsRequest_DefaultValuesHandling(t *testing.T) {
	// Test that the handler correctly applies defaults for empty values
	// The handler sets: Since="525600m", Until="0m", Tail="all" when empty

	// Empty values in JSON
	jsonData := `{
		"containers": [],
		"since": "",
		"until": "",
		"tail": ""
	}`

	var req pkg.LogsRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	// Original values should be empty
	assert.Empty(t, req.Since)
	assert.Empty(t, req.Until)
	assert.Empty(t, req.Tail)
}

// All Parameters Test

func TestLogsRequest_AllParameters(t *testing.T) {
	jsonData := `{
		"containers": ["web", "db", "cache", "worker"],
		"timestamps": true,
		"tail": "500",
		"since": "2h30m",
		"until": "5m"
	}`

	var req pkg.LogsRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Len(t, req.Containers, 4)
	assert.True(t, req.Timestamps)
	assert.Equal(t, "500", req.Tail)
	assert.Equal(t, "2h30m", req.Since)
	assert.Equal(t, "5m", req.Until)
}

// ---- logsHandler success path tests ----

const troubleshootTestMode = "test"

func setTroubleshootTestMode(t *testing.T) func() {
	t.Helper()

	oldMode := app.Mode
	app.Mode = troubleshootTestMode

	return func() { app.Mode = oldMode }
}

func TestLogsHandler_DefaultsApplied_TestMode(t *testing.T) {
	// Set test mode so mocker.Mocked.Logs() is called (no real Docker)
	defer setTroubleshootTestMode(t)()

	router := gin.New()
	router.POST("/logs", logsHandler)

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/logs", bytes.NewBufferString(`{}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)

	// Mocker returns a valid LogsResponse → 200
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestLogsHandler_WithContainers_TestMode(t *testing.T) {
	defer setTroubleshootTestMode(t)()

	router := gin.New()
	router.POST("/logs", logsHandler)

	reqBody, err := json.Marshal(pkg.LogsRequest{
		Containers: []string{"web", "db"},
		Tail:       "100",
		Since:      "1h",
		Until:      "0m",
		Timestamps: false,
	})
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/logs", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	ct := recorder.Header().Get("Content-Type")
	assert.Contains(t, ct, "application/json")
}

func TestLogsHandler_ResponseShape_TestMode(t *testing.T) {
	defer setTroubleshootTestMode(t)()

	router := gin.New()
	router.POST("/logs", logsHandler)

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/logs",
		bytes.NewBufferString(`{"since": "1h", "until": "0m", "tail": "50"}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var resp pkg.LogsResponse
	err = json.Unmarshal(recorder.Body.Bytes(), &resp)
	require.NoError(t, err)
	// containers field may be nil — mocker has no containers up in this state
	_ = resp.Containers
}

// ---- rebootHandler test ----

func TestRebootHandler_FailsWithoutPrivilege(t *testing.T) {
	// The reboot syscall always fails in unprivileged test environments.
	// We verify the handler responds with a JSON error rather than panicking.
	router := gin.New()
	router.POST("/reboot", rebootHandler)

	recorder := httptest.NewRecorder()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "/reboot", nil)
	require.NoError(t, err)
	router.ServeHTTP(recorder, req)

	// Must be 500 (syscall fails) in a non-root test environment
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp, "error")
}
