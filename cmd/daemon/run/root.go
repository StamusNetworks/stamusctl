package run

import (
	// Common
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	// Custom
	docs "stamus-ctl/cmd/daemon/docs"
	"stamus-ctl/cmd/daemon/run/compose"
	"stamus-ctl/cmd/daemon/run/config"
	"stamus-ctl/cmd/daemon/run/health"
	"stamus-ctl/cmd/daemon/run/troubleshoot"
	"stamus-ctl/internal/auth"
	"stamus-ctl/internal/logging"
	"stamus-ctl/internal/middleware"

	// External

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
)

const (
	// RateLimitPerSecond defines the number of requests allowed per second per IP
	RateLimitPerSecond = 5
)


// Ping godoc
// @Summary ping example
// @Schemes
// @Description do ping
// @Tags example
// @Accept json
// @Produce json
// @Success 200 {string} Helloworld
// @Router /ping [get]
// @Security BasicAuth
func ping(c *gin.Context) {
	logging.LoggerWithRequest(c.Request).Info("Ping")

	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func RunCmd() *cobra.Command {
	_, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Initialize OpenTelemetry tracing
	collectorURL := getEnvFallback("OTEL_COLLECTOR_URL", "")
	serviceName := getEnvFallback("OTEL_SERVICE_NAME", "stamus-ctl-daemon")
	logging.InitTracer(collectorURL, serviceName)

	viper.SetDefault("tokenpath", "")

	viper.SetEnvPrefix("stamusd")
	viper.BindEnv("tokenpath")

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			span := setupLogging()
			logger := getLogger(span)
			r := SetupRouter(logger)
			logger("Starting daemon")
			err := r.Run(":8080")
			if err != nil {
				logger(err.Error())
				return err
			}
			return nil
		},
	}

	return cmd
}

func setupLogging() trace.Span {
	c := context.Background()
	ctx, span := logging.Tracer.Start(c, "main")
	defer span.End()
	go logging.NewPrometheusServer(ctx)
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

	if viper.GetString("tokenpath") != "" {
		go auth.WatchForToken(viper.GetString("tokenpath"))
	}

	// Health endpoints (no auth required for Kubernetes probes)
	logger("Setup health endpoints")
	health.NewHealth(r)

	// Middleware
	r.Use(gin.Recovery())
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware())
	if viper.GetString("tokenpath") != "" {
		r.Use(otelgin.Middleware("stamusd", otelgin.WithTracerProvider(logging.TracerProvider)))
		r.Use(logging.SetRequestIDInResponse())
		r.Use(logging.LogRequestResponse())
	}
	r.Use(auth.AuthMiddleware())

	// Routes
	logger("Setup routes")
	v1 := r.Group("/api/v1")
	v1.Use(RateLimiter())
	{
		v1.GET("/ping", ping)
		v1.POST("/upload", uploadHandler)
		compose.NewCompose(v1)
		config.NewConfig(v1)
		troubleshoot.NewTroubleshoot(v1)
	}

	// Swagger
	logger("Setup swagger")
	docs.SwaggerInfo.BasePath = "/api/v1"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// r.RunUnix("./daemon.sock")
	return r
}

func getEnvFallback(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return fallback
}

func errorHandler(c *gin.Context, info ratelimit.Info) {
	c.String(http.StatusTooManyRequests, "Too many requests. Try again in "+time.Until(info.ResetTime).String())
}

func keyFunc(c *gin.Context) string {
	return c.ClientIP()
}

// RateLimiter returns a gin middleware that limits the number of requests per IP address.
// It attempts to use Redis for distributed rate limiting, but falls back to in-memory
// rate limiting if Redis is unavailable.
func RateLimiter() gin.HandlerFunc {
	redisHost := getEnvFallback("REDIS_HOST", "localhost")
	redisPort := getEnvFallback("REDIS_PORT", "6379")
	redisPass := getEnvFallback("REDIS_PASSWORD", "license")

	// Each ip can make 5 requests per second
	client := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPass,
		DB:       0, // use default DB
	})

	// Test Redis connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var store ratelimit.Store
	if err := client.Ping(ctx).Err(); err != nil {
		// Redis unavailable - fall back to in-memory store
		logging.Logger.Warn("Redis unavailable, falling back to in-memory rate limiter: " + err.Error())
		store = ratelimit.InMemoryStore(&ratelimit.InMemoryOptions{
			Rate:  time.Second,
			Limit: RateLimitPerSecond,
		})
	} else {
		// Redis available - use distributed rate limiting
		logging.Logger.Info("Using Redis for distributed rate limiting at " + redisHost + ":" + redisPort)
		store = ratelimit.RedisStore(&ratelimit.RedisOptions{
			RedisClient: client,
			Rate:        time.Second,
			Limit:       RateLimitPerSecond,
		})
	}

	return ratelimit.RateLimiter(store, &ratelimit.Options{
		ErrorHandler: errorHandler,
		KeyFunc:      keyFunc,
	})
}
