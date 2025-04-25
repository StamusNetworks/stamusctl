package compose

import (
	// Internal
	handlers "stamus-ctl/internal/handlers/compose"
	"stamus-ctl/pkg"

	// External
	"github.com/gin-gonic/gin"
)

// restartConfigHandler godoc
// @Summary Similar to docker restart
// @Description Will restart the containers defined in the current configuration
// @Tags compose
// @Accept json
// @Produce json
// @Success 200 {object} pkg.SuccessResponse "Success message"
// @Failure 400 {object} pkg.ErrorResponse "Bad request with explanation"
// @Router /compose/restart/config [post]
func restartConfigHandler(c *gin.Context) {
	// Call handler
	err := handlers.HandleConfigRestart(c.Query("config"))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

// RestartContainerHandler godoc
// @Summary Similar to docker restart
// @Description Will restart the containers defined
// @Tags compose
// @Accept json
// @Produce json
// @Param arbitraries body pkg.Containers true "Containers to restart"
// @Success 200 {object} pkg.SuccessResponse "Success message"
// @Failure 400 {object} pkg.ErrorResponse "Bad request with explanation"
// @Router /compose/restart/containers [post]
func restartContainersHandler(c *gin.Context) {
	// Extract request body
	var req pkg.Containers
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.Containers == nil {
		req.Containers = []string{}
	}

	// Call handler
	err := handlers.HandleContainersRestart(req.Containers)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "restart successful"})
}
