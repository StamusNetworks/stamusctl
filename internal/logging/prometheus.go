package logging

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

// NewPrometheusServer starts a Prometheus metrics server on port 9001.
// It shuts down gracefully when the provided context is cancelled.
func NewPrometheusServer(ctx context.Context) {
	engineProm := gin.New()

	p := ginprometheus.NewPrometheus("gin")
	p.Use(engineProm)

	srv := &http.Server{
		Addr:    ":9001",
		Handler: engineProm,
	}

	// Start server in goroutine
	go func() {
		LoggerWithContextToSpanContext(ctx).Info("Starting prometheus endpoint on :9001")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			Logger.Error("Prometheus server error: " + err.Error())
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	LoggerWithContextToSpanContext(ctx).Info("Shutting down prometheus server")

	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*1e9) // 5 seconds
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		Logger.Error("Prometheus server shutdown error: " + err.Error())
	}
}
