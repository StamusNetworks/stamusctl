package compose

import (
	// Internal
	handlers "stamus-ctl/internal/handlers/wrapper"

	// External
	"github.com/gin-gonic/gin"
)

// UpHandler godoc
// @Summary Similar to docker compose up
// @Description Starts the services defined in the current configuration.
// @Tags compose
// @Accept json
// @Produce json
// @Success 200 {object} pkg.SuccessResponse "Up successful"
// @Failure 400 {object} pkg.ErrorResponse "Bad request with explanation"
// @Router /compose/up [post]
func upHandler(c *gin.Context) {
	// Call handler
	err := handlers.HandleUp()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

// UpHandler godoc
// @Summary Similar to docker compose down
// @Description Stops the services defined in the current configuration.
// @Tags compose
// @Accept json
// @Produce json
// @Success 200 {object} pkg.SuccessResponse "Down successful"
// @Failure 400 {object} pkg.ErrorResponse "Bad request with explanation"
// @Router /compose/down [post]
func downHandler(c *gin.Context) {
	// Call handler
	err := handlers.HandleDown(false, false)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}
