package troubleshoot

import (
	"github.com/docker/docker/pkg/dmesg"
	"github.com/gin-gonic/gin"
)

// kernelHandler godoc
// @Summary Logs of the kernel
// @Description Will return the logs of the kernel
// @Tags logs
// @Produce json
// @Success 200 {object} pkg.SuccessResponse "Kernel logs"
// @Failure 500 {object} pkg.ErrorResponse "Internal server error with explanation"
// @Router /troubleshoot/kernel [post]
func kernelHandler(c *gin.Context) {
	msg := dmesg.Dmesg(32768)
	c.JSON(200, gin.H{"message": msg})
}
