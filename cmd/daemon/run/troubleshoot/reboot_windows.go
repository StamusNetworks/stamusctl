package troubleshoot

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// rebootHandler godoc
// @Summary Reboots the system
// @Description Will reboot the system
// @Tags reboot
// @Failure 500 {object} pkg.ErrorResponse "Internal server error with explanation"
// @Router /troubleshoot/reboot [post]
func rebootHandler(c *gin.Context) {
	// Reboot the system
	c.JSON(500, gin.H{"error": fmt.Sprintf("failed to reboot: we are on windows")})
	return
}
