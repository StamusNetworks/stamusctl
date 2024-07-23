package run

import (
	// Common
	"context"
	"os"
	"os/signal"

	// Custom
	"stamus-ctl/cmd/daemon/run/compose"
	"stamus-ctl/cmd/daemon/run/config"
	docs "stamus-ctl/cmd/docs"
	"stamus-ctl/internal/logging"

	// External
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	ginprometheus "github.com/zsais/go-gin-prometheus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

func NewPrometheusServer(ctx context.Context) {
	engineProm := gin.New()

	p := ginprometheus.NewPrometheus("gin")
	p.Use(engineProm)

	logging.LoggerWithContextToSpanContext(ctx).Info("Starting prometheus endpoint")
	engineProm.Run(":9001")
}

// Ping godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /ping [get]
func ping(c *gin.Context) {
	logging.LoggerWithRequest(c.Request).Info("Ping")

	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// @title           Swagger Stamusd API
// @version         1.0
// @description     Stamus daemon server.

// @BasePath /api/v1
func RunCmd() *cobra.Command {
	_, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	logging.NewTraceProvider()

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			span := setupLogging()
			logger := getLogger(span)
			r := SetupRouter(logger)
			logger("Starting daemon")
			r.Run(":8080")
			return nil
		},
	}

	return cmd
}

func setupLogging() trace.Span {
	c := context.Background()
	ctx, span := logging.Tracer.Start(c, "main")
	defer span.End()
	go NewPrometheusServer(ctx)
	return span
}

func getLogger(span trace.Span) func(string) {
	return func(message string) {
		logging.LoggerWithSpanContext(span.SpanContext()).Info(message)
	}
}

func SetupRouter(logger func(string)) *gin.Engine {
	// Gin setup
	logger("Setup middleware")
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// Middleware
	r.Use(gin.Recovery())
	r.Use(otelgin.Middleware("stamusd", otelgin.WithTracerProvider(logging.TracerProvider)))

	// Routes
	logger("Setup routes")
	v1 := r.Group("/api/v1")
	{
		v1.GET("/ping", ping)
		v1.POST("/upload", uploadHandler)
		compose.NewCompose(v1)
		config.NewConfig(v1)
	}

	// Swagger
	logger("Setup swagger")
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// r.RunUnix("./daemon.sock")
	return r

}
