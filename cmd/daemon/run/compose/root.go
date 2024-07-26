package compose

import "github.com/gin-gonic/gin"

// Ping godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /compose/pingcompose [get]
func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pongcompose",
	})
}

func NewCompose(router *gin.RouterGroup) {
	compose := router.Group("/compose")
	{
		compose.GET("/pingcompose", ping)
		compose.POST("/init", initHandler)
		compose.POST("/update", updateHandler)
		compose.POST("/up", upHandler)
		compose.POST("/ps", psHandler)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
