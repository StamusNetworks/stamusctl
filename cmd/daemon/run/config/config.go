package config

import (
	// External
	"github.com/gin-gonic/gin"

	// Internal
	"stamus-ctl/internal/app"
	handlers "stamus-ctl/internal/handlers/config"
	"stamus-ctl/internal/logging"
	"stamus-ctl/pkg"
)

// setHandler godoc
// @Summary Set configuration
// @Description Sets configuration with provided parameters.
// @Tags config
// @Accept json
// @Produce json
// @Param set body pkg.SetRequest true "Set parameters"
// @Success 200 {object} pkg.SuccessResponse "Configuration set successfully"
// @Failure 400 {object} pkg.ErrorResponse "Bad request with explanation"
// @Failure 500 {object} pkg.ErrorResponse "Internal server error with explanation"
// @Router /config [post]
func setHandler(c *gin.Context) {
	// Extract request body
	var req pkg.SetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	conf := req.Config
	if conf == "" {
		conf = app.DefaultConfigName
	}
	if req.Values == nil {
		req.Values = make(map[string]string)
	}
	valuesAsStrings := []string{}
	for k, v := range req.Values {
		valuesAsStrings = append(valuesAsStrings, k+"="+v)
	}
	fromFile := ""
	if req.FromFile != nil {
		for k, v := range req.FromFile {
			fromFile += k + "=" + v + " "
		}
	}

	// Call handler
	params := handlers.SetHandlerInputs{
		Reload:   req.Reload,
		Apply:    req.Apply,
		Args:     valuesAsStrings,
		Values:   req.ValuesPath,
		FromFile: fromFile,
		Config:   app.GetConfigsFolder(conf),
	}
	if err := handlers.SetHandler(params); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "ok"})
}

// getConfigListHandler godoc
// @Summary Set current configuration
// @Description Sets configuration with provided parameters.
// @Tags config
// @Accept json
// @Produce json
// @Success 200 {object} pkg.GetListResponse "Configuration list"
// @Failure 500 {object} pkg.ErrorResponse "Internal server error with explanation"
// @Router /config/list [post]
func getConfigListHandler(c *gin.Context) {
	// Call handler
	list, err := handlers.GetConfigsList()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	resp := pkg.GetListResponse{Configs: list}
	c.JSON(200, resp)
}

type GetResponse map[string]interface{}

// getHandler godoc
// @Summary Get configuration
// @Description Retrieves configuration for a given project.
// @Tags config
// @Produce json
// @Param get query pkg.GetRequest true "Get parameters"
// @Success 200 {object} GetResponse "Configuration retrieved successfully"
// @Failure 404 {object} pkg.ErrorResponse "Bad request with explanation"
// @Failure 500 {object} pkg.ErrorResponse "Internal server error with explanation"
// @Router /config [get]
func getHandler(c *gin.Context) {
	// Extract request body
	var req pkg.GetRequest
	if err := c.BindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// Validate request
	if req.Values == nil {
		req.Values = []string{}
	}

	conf := req.Config
	if conf == "" {
		conf = app.DefaultConfigName
	}
	// Call handler
	if req.Content {
		filesMap, err := handlers.GetGroupedContent(conf, req.Values)
		if err != nil {
			logging.LoggerWithRequest(c.Request).Error(err.Error())
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, filesMap)
		return
	}

	groupedValues, err := handlers.GetGroupedConfig(conf, req.Values, false)
	if err != nil {
		logging.LoggerWithRequest(c.Request).Error(err.Error())
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, groupedValues)
}

type VersionResponse string

// versionHandler godoc
// @Summary Get configuration version
// @Description Retrieves configuration version for a given project.
// @Tags config
// @Produce json
// @Param get query pkg.Config true "Config to get version from"
// @Success 200 {object} VersionResponse "Configuration version retrieved successfully"
// @Failure 404 {object} pkg.ErrorResponse "Bad request with explanation"
// @Failure 500 {object} pkg.ErrorResponse "Internal server error with explanation"
// @Router /config [get]
func getVersionHandler(c *gin.Context) {
	// Extract request body
	var req pkg.Config
	if err := c.BindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	version := handlers.GetVersion(req.Value)
	c.JSON(200, version)
}
