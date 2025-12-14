package logging

import (
	"context"

	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

// NewPrometheusServer starts a Prometheus metrics server on port 9001.
func NewPrometheusServer(ctx context.Context) {
	engineProm := gin.New()

	p := ginprometheus.NewPrometheus("gin")
	p.Use(engineProm)

	LoggerWithContextToSpanContext(ctx).Info("Starting prometheus endpoint")
	_ = engineProm.Run(":9001")
}
