package health

import (
	"github.com/gin-gonic/gin"
)

// NewHealth registers health check endpoints
func NewHealth(router *gin.Engine) {
	// Health and readiness checks should be at root level for Kubernetes
	// and not require authentication
	router.GET("/health", healthHandler)
	router.GET("/ready", readinessHandler)
}
