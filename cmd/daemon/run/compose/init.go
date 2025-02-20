package compose

import (
	// External
	"github.com/gin-gonic/gin"

	// Custom
	"stamus-ctl/internal/app"
	handlers "stamus-ctl/internal/handlers/compose"
	"stamus-ctl/pkg"
)

// InitHandler godoc
// @Summary Initialize configuration
// @Description Initializes configuration with provided parameters.
// @Tags compose
// @Accept json
// @Produce json
// @Param arbitraries body pkg.InitRequest true "Initialization parameters"
// @Success 200 {object} pkg.SuccessResponse "Initialization successful"
// @Failure 400 {object} pkg.ErrorResponse "Bad request with explanation"
// @Router /compose/init [post]
func initHandler(c *gin.Context) {
	// Extract request body
	var req pkg.InitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.Values == nil {
		req.Values = make(map[string]string)
	}
	if req.Config == "" {
		req.Config = "config"
	}
	if req.Project == "" {
		req.Project = "clearndr"
	}
	if req.Version == "" {
		req.Version = "latest"
	}
	fromFile := ""
	if req.FromFile != nil {
		for k, v := range req.FromFile {
			fromFile += k + "=" + v + " "
		}
	}

	// Call handler
	parameters := handlers.InitHandlerInputs{
		IsDefault:        req.IsDefault,
		BackupFolderPath: app.DefaultClearNDRPath,
		Arbitrary:        req.Values,
		Project:          req.Project,
		Version:          req.Version,
		Values:           req.ValuesPath,
		Registry:         req.Registry,
		FromFile:         fromFile,
		Config:           app.GetConfigsFolder(req.Config),
	}
	if err := handlers.InitHandler(false, parameters); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})

}
