package troubleshoot

import (
	"fmt"
	"syscall"

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
	err := syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to reboot: %v", err)})
		return
	}
}
