package run

import (
	// Custom

	"os"
	"path/filepath"
	"stamus-ctl/internal/logging"

	// External
	"github.com/gin-gonic/gin"
)

// UploadHandler godoc
// @Summary Upload file example
// @Schemes
// @Description Handles file uploads
// @Tags example
// @Accept multipart/form-data
// @Produce json
// @Param path query string true "Path to save file"
// @Param project query string false "Project name"
// @Param file formData file true "Upload file"
// @Success 200 {string} string "Uploaded file"
// @Failure 400 {string} string "Error message"
// @Failure 500 {string} string "Error message"
// @Router /upload [post]
func uploadHandler(c *gin.Context) {
	logging.LoggerWithRequest(c.Request).Info("Upload")

	// Validate request
	if c.Request.ContentLength == 0 {
		c.String(400, "No file uploaded")
		return
	}
	if c.Query("path") == "" {
		c.String(400, "No path provided")
		return
	}
	project := c.Query("project")

	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		c.String(400, "File upload error: "+err.Error())
		return
	}

	// Create directory if it doesn't exist
	err = os.MkdirAll(filepath.Join(project, c.Query("path")), 0755)
	if err != nil {
		c.String(500, "Directory creation error: "+err.Error())
		return
	}

	// Save file to path
	err = c.SaveUploadedFile(file, filepath.Join(project, c.Query("path"), file.Filename))
	if err != nil {
		c.String(500, "File save error: "+err.Error())
		return
	}

	c.String(200, "Uploaded file")
}
