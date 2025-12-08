package health

import (
	"context"
	"os"
	"time"

	"stamus-ctl/internal/app"
	"stamus-ctl/internal/logging"
	"stamus-ctl/pkg"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

// Readiness godoc
// @Summary Readiness check endpoint
// @Description Returns readiness status of the daemon with detailed checks (readiness probe for Kubernetes)
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} pkg.ReadinessResponse
// @Failure 503 {object} pkg.ReadinessResponse
// @Router /ready [get]
func readinessHandler(c *gin.Context) {
	logging.LoggerWithRequest(c.Request).Info("Readiness check")

	checks := make(map[string]string)
	allReady := true

	// Check Docker daemon connectivity
	dockerConnected := checkDockerConnectivity()
	if dockerConnected {
		checks["docker_daemon"] = "connected"
	} else {
		checks["docker_daemon"] = "disconnected"
		allReady = false
	}

	// Check configuration validity
	configValid := checkConfigurationValidity()
	if configValid {
		checks["configuration"] = "valid"
	} else {
		checks["configuration"] = "not_found"
		// Configuration not being present is not necessarily a blocker for readiness
		// as the daemon can still accept requests to initialize configuration
		// So we don't set allReady to false here
	}

	// Check required resources (config folder exists and is writable)
	resourcesAvailable := checkRequiredResources()
	if resourcesAvailable {
		checks["resources"] = "available"
	} else {
		checks["resources"] = "unavailable"
		allReady = false
	}

	response := pkg.ReadinessResponse{
		Checks:          checks,
		DockerConnected: dockerConnected,
	}

	if allReady {
		response.Status = "ready"
		response.Message = "daemon is ready to accept requests"
		c.JSON(200, response)
	} else {
		response.Status = "not_ready"
		response.Message = "daemon is not ready, check individual checks for details"
		c.JSON(503, response)
	}
}

// checkDockerConnectivity verifies that the Docker daemon is accessible
func checkDockerConnectivity() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return false
	}
	defer apiClient.Close()

	// Try to ping the Docker daemon
	_, err = apiClient.Ping(ctx)
	if err != nil {
		// Fallback: try to list containers as a connectivity check
		_, err = apiClient.ContainerList(ctx, container.ListOptions{Limit: 1})
		if err != nil {
			return false
		}
	}

	return true
}

// checkConfigurationValidity checks if a valid configuration exists
func checkConfigurationValidity() bool {
	// Check if the default config folder exists
	configPath := app.GetConfigsFolder("config")
	if _, err := os.Stat(configPath); err == nil {
		return true
	}

	// If default config doesn't exist, that's okay - the daemon can still accept
	// requests to create a new configuration
	return false
}

// checkRequiredResources verifies that required directories exist and are writable
func checkRequiredResources() bool {
	// Check if the base configs folder exists and is writable
	configsFolder := app.ConfigsFolder
	if info, err := os.Stat(configsFolder); err != nil {
		// Try to create it if it doesn't exist
		if os.IsNotExist(err) {
			if err := os.MkdirAll(configsFolder, 0o755); err != nil {
				return false
			}
		} else {
			return false
		}
	} else if !info.IsDir() {
		return false
	}

	// Test write permissions by attempting to create a temp file
	testFile := configsFolder + "/.healthcheck"
	if err := os.WriteFile(testFile, []byte("test"), 0o644); err != nil {
		return false
	}
	os.Remove(testFile)

	return true
}
