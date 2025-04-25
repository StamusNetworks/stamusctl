package compose

import (
	handlers "stamus-ctl/internal/handlers/compose"
	"stamus-ctl/pkg"

	"github.com/gin-gonic/gin"
)

// Init godoc
// @Summary Update configuration
// @Description Update configuration with provided parameters.
// @Tags compose
// @Accept json
// @Produce json
// @Param arbitraries body pkg.UpdateRequest true "Update parameters"
// @Success 200 {object} pkg.SuccessResponse "Initialization successful"
// @Failure 400 {object} pkg.ErrorResponse "Bad request with explanation"
// @Router /compose/update [post]
func updateHandler(c *gin.Context) {
	// Extract request body
	var req pkg.UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Validate parameters
	if req.Version == "" {
		req.Version = "latest"
	}
	if req.Values == nil {
		req.Values = make(map[string]string)
	}
	valuesVal := []string{}
	for k, v := range req.Values {
		valuesVal = append(valuesVal, k+"="+v)
	}

	// Call handler
	params := handlers.UpdateHandlerParams{
		Version: req.Version,
		Args:    valuesVal,
	}
	err := handlers.UpdateHandler(params)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Configuration updated successfully"})
}
