package health

import (
	"stamus-ctl/internal/logging"
	"stamus-ctl/pkg"

	"github.com/gin-gonic/gin"
)

// Health godoc
// @Summary Health check endpoint
// @Description Returns health status of the daemon (liveness probe for Kubernetes)
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} pkg.HealthResponse
// @Router /health [get]
func healthHandler(c *gin.Context) {
	logging.LoggerWithRequest(c.Request).Info("Health check")

	c.JSON(200, pkg.HealthResponse{
		Status:  "ok",
		Message: "daemon is running",
	})
}
