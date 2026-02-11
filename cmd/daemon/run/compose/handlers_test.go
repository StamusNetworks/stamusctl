package compose

import (
	"bytes"
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

// Route Registration Tests

func TestNewCompose_RegistersRoutes(t *testing.T) {
	router := gin.New()
	v1 := router.Group("/api/v1")
	NewCompose(v1)

	routes := router.Routes()

	expectedRoutes := map[string]string{
		"/api/v1/compose/init":               http.MethodPost,
		"/api/v1/compose/update":             http.MethodPost,
		"/api/v1/compose/up":                 http.MethodPost,
		"/api/v1/compose/down":               http.MethodPost,
		"/api/v1/compose/ps":                 http.MethodPost,
		"/api/v1/compose/restart/config":     http.MethodPost,
		"/api/v1/compose/restart/containers": http.MethodPost,
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

// JSON Parsing Tests - These only test that handlers correctly parse JSON
// without actually executing the Docker operations

func TestInitHandler_InvalidJSON(t *testing.T) {
	router := gin.New()
	router.POST("/init", initHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/init", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestUpdateHandler_InvalidJSON(t *testing.T) {
	router := gin.New()
	router.POST("/update", updateHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/update", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestRestartContainersHandler_InvalidJSON(t *testing.T) {
	router := gin.New()
	router.POST("/restart", restartContainersHandler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/restart", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "error")
}

// Request Struct Tests - Verify request structs are correctly defined

func TestInitRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"config": "myconfig",
		"default": true,
		"project": "myproject",
		"values": {"key1": "val1"},
		"version": "v1.0.0",
		"values_path": "/path/to/values.yaml",
		"registry": "docker.io"
	}`

	var req pkg.InitRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Equal(t, "myconfig", req.Config)
	assert.True(t, req.IsDefault)
	assert.Equal(t, "myproject", req.Project)
	assert.Equal(t, "val1", req.Values["key1"])
	assert.Equal(t, "v1.0.0", req.Version)
	assert.Equal(t, "/path/to/values.yaml", req.ValuesPath)
	assert.Equal(t, "docker.io", req.Registry)
}

func TestUpdateRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"version": "v2.0.0",
		"values": {"key1": "newval1", "key2": "newval2"}
	}`

	var req pkg.UpdateRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Equal(t, "v2.0.0", req.Version)
	assert.Equal(t, "newval1", req.Values["key1"])
	assert.Equal(t, "newval2", req.Values["key2"])
}

func TestContainersRequest_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"containers": ["container1", "container2", "container3"]
	}`

	var req pkg.Containers
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Len(t, req.Containers, 3)
	assert.Contains(t, req.Containers, "container1")
	assert.Contains(t, req.Containers, "container2")
	assert.Contains(t, req.Containers, "container3")
}

func TestInitRequest_EmptyJSON(t *testing.T) {
	jsonData := `{}`

	var req pkg.InitRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	// Should have zero values
	assert.Empty(t, req.Config)
	assert.False(t, req.IsDefault)
	assert.Empty(t, req.Project)
	assert.Nil(t, req.Values)
	assert.Empty(t, req.Version)
}

func TestUpdateRequest_EmptyJSON(t *testing.T) {
	jsonData := `{}`

	var req pkg.UpdateRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Empty(t, req.Version)
	assert.Nil(t, req.Values)
}

func TestContainersRequest_EmptyJSON(t *testing.T) {
	jsonData := `{}`

	var req pkg.Containers
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Nil(t, req.Containers)
}

// Method Tests

func TestComposeRoutes_OnlyAcceptPost(t *testing.T) {
	router := gin.New()
	v1 := router.Group("/api/v1")
	NewCompose(v1)

	endpoints := []string{
		"/api/v1/compose/init",
		"/api/v1/compose/update",
		"/api/v1/compose/up",
		"/api/v1/compose/down",
		"/api/v1/compose/ps",
		"/api/v1/compose/restart/config",
		"/api/v1/compose/restart/containers",
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

// JSON Marshal/Unmarshal roundtrip tests

func TestInitRequest_MarshalUnmarshal(t *testing.T) {
	original := pkg.InitRequest{
		Config:     "testconfig",
		IsDefault:  true,
		Project:    "testproject",
		Values:     map[string]string{"key": "value"},
		Version:    "1.0.0",
		ValuesPath: "/path/to/values.yaml",
		Registry:   "registry.example.com",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)

	// Unmarshal back
	var decoded pkg.InitRequest
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Config, decoded.Config)
	assert.Equal(t, original.IsDefault, decoded.IsDefault)
	assert.Equal(t, original.Project, decoded.Project)
	assert.Equal(t, original.Values, decoded.Values)
	assert.Equal(t, original.Version, decoded.Version)
	assert.Equal(t, original.ValuesPath, decoded.ValuesPath)
	assert.Equal(t, original.Registry, decoded.Registry)
}

func TestUpdateRequest_MarshalUnmarshal(t *testing.T) {
	original := pkg.UpdateRequest{
		Version: "2.0.0",
		Values:  map[string]string{"newkey": "newvalue"},
	}

	jsonData, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded pkg.UpdateRequest
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Version, decoded.Version)
	assert.Equal(t, original.Values, decoded.Values)
}

func TestContainersRequest_MarshalUnmarshal(t *testing.T) {
	original := pkg.Containers{
		Containers: []string{"web", "db", "cache"},
	}

	jsonData, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded pkg.Containers
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.Containers, decoded.Containers)
}

// FromFile parsing tests

func TestInitRequest_WithFromFile(t *testing.T) {
	jsonData := `{
		"from_file": {
			"cert": "/path/to/cert.pem",
			"key": "/path/to/key.pem"
		}
	}`

	var req pkg.InitRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/cert.pem", req.FromFile["cert"])
	assert.Equal(t, "/path/to/key.pem", req.FromFile["key"])
}
