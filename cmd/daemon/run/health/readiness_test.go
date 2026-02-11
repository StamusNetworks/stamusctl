package health

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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

// ReadinessResponse struct tests (no Docker calls)

func TestReadinessResponse_JSONMarshal(t *testing.T) {
	response := pkg.ReadinessResponse{
		Status:          "ready",
		Message:         "daemon is ready to accept requests",
		DockerConnected: true,
		Checks: map[string]string{
			"docker_daemon": "connected",
			"configuration": "valid",
			"resources":     "available",
		},
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var decoded pkg.ReadinessResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "ready", decoded.Status)
	assert.Equal(t, "daemon is ready to accept requests", decoded.Message)
	assert.True(t, decoded.DockerConnected)
	assert.Equal(t, "connected", decoded.Checks["docker_daemon"])
	assert.Equal(t, "valid", decoded.Checks["configuration"])
	assert.Equal(t, "available", decoded.Checks["resources"])
}

func TestReadinessResponse_NotReadyState(t *testing.T) {
	response := pkg.ReadinessResponse{
		Status:          "not_ready",
		Message:         "daemon is not ready, check individual checks for details",
		DockerConnected: false,
		Checks: map[string]string{
			"docker_daemon": "disconnected",
			"configuration": "not_found",
			"resources":     "unavailable",
		},
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var decoded pkg.ReadinessResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "not_ready", decoded.Status)
	assert.False(t, decoded.DockerConnected)
	assert.Equal(t, "disconnected", decoded.Checks["docker_daemon"])
}

func TestReadinessResponse_EmptyChecks(t *testing.T) {
	response := pkg.ReadinessResponse{
		Status:  "not_ready",
		Message: "no checks performed",
		Checks:  map[string]string{},
	}

	jsonData, err := json.Marshal(response)
	require.NoError(t, err)

	var decoded pkg.ReadinessResponse
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.NotNil(t, decoded.Checks)
	assert.Empty(t, decoded.Checks)
}

func TestNewHealth_RegistersReadyRoute(t *testing.T) {
	router := gin.New()
	NewHealth(router)

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

func TestCheckConfigurationValidity_ExistingConfig(t *testing.T) {
	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Create a temp directory for testing with real filesystem
	tempDir := t.TempDir()
	app.ConfigsFolder = tempDir + "/"

	// Create the config subdirectory using real OS
	configDir := tempDir + "/config"
	err := os.MkdirAll(configDir, 0755)
	require.NoError(t, err)

	result := checkConfigurationValidity()
	assert.True(t, result)
}

func TestCheckConfigurationValidity_MissingConfig(t *testing.T) {
	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Create a temp directory for testing with real filesystem
	tempDir := t.TempDir()
	app.ConfigsFolder = tempDir + "/"
	// Don't create the "config" subdirectory

	result := checkConfigurationValidity()
	assert.False(t, result)
}

func TestCheckRequiredResources_WritableDir(t *testing.T) {
	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Create a temp directory for testing with real filesystem
	tempDir := t.TempDir()
	app.ConfigsFolder = tempDir + "/"

	result := checkRequiredResources()
	assert.True(t, result)
}

func TestCheckRequiredResources_CreateDir(t *testing.T) {
	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Create a temp directory and point to a non-existent subdirectory
	tempDir := t.TempDir()
	newConfigsDir := filepath.Join(tempDir, "new-configs")
	app.ConfigsFolder = newConfigsDir

	result := checkRequiredResources()
	assert.True(t, result)

	// Verify directory was created using real OS
	info, err := os.Stat(newConfigsDir)
	assert.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestCheckRequiredResources_NotWritable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping test as root user can write to read-only directories")
	}

	// Create a temp directory
	tempDir := t.TempDir()
	readOnlyDir := filepath.Join(tempDir, "readonly")

	// Create the directory first
	err := os.MkdirAll(readOnlyDir, 0755)
	require.NoError(t, err)

	// Make it read-only
	err = os.Chmod(readOnlyDir, 0444)
	require.NoError(t, err)
	defer os.Chmod(readOnlyDir, 0755) // Restore permissions for cleanup

	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	app.ConfigsFolder = readOnlyDir

	result := checkRequiredResources()
	assert.False(t, result)
}

func TestCheckRequiredResources_NotADir(t *testing.T) {
	// Save original ConfigsFolder
	originalConfigsFolder := app.ConfigsFolder
	defer func() { app.ConfigsFolder = originalConfigsFolder }()

	// Create a temp directory and create a file instead of directory
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "configs-file")
	err := os.WriteFile(filePath, []byte("not a directory"), 0644)
	require.NoError(t, err)

	app.ConfigsFolder = filePath

	result := checkRequiredResources()
	assert.False(t, result)
}

// Additional struct validation tests

func TestReadinessResponse_ChecksMapOperations(t *testing.T) {
	response := pkg.ReadinessResponse{
		Status: "ready",
		Checks: make(map[string]string),
	}

	// Add checks dynamically
	response.Checks["docker_daemon"] = "connected"
	response.Checks["configuration"] = "valid"
	response.Checks["resources"] = "available"

	assert.Len(t, response.Checks, 3)
	assert.Equal(t, "connected", response.Checks["docker_daemon"])
}

func TestReadinessResponse_StatusValues(t *testing.T) {
	validStatuses := []string{"ready", "not_ready"}

	for _, status := range validStatuses {
		response := pkg.ReadinessResponse{Status: status}
		assert.Contains(t, validStatuses, response.Status)
	}
}

func TestReadinessResponse_MessageValues(t *testing.T) {
	readyMsg := "daemon is ready to accept requests"
	notReadyMsg := "daemon is not ready, check individual checks for details"

	readyResp := pkg.ReadinessResponse{Status: "ready", Message: readyMsg}
	assert.Equal(t, readyMsg, readyResp.Message)

	notReadyResp := pkg.ReadinessResponse{Status: "not_ready", Message: notReadyMsg}
	assert.Equal(t, notReadyMsg, notReadyResp.Message)
}
