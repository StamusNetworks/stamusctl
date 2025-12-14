package logging

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// SetRequestIDInResponse creates middleware that adds the trace ID as X-Request-ID header to responses.
func SetRequestIDInResponse() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r := ctx.Request
		span := trace.SpanContextFromContext(r.Context())
		ctx.Writer.Header().Set("X-Request-ID", span.TraceID().String())
		ctx.Next()
	}
}

// LogRequestResponse creates middleware that logs incoming requests and outgoing responses.
func LogRequestResponse() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		r := ctx.Request
		LoggerWithRequest(r).Info("Request", zap.String("url", r.URL.String()), zap.String("method", r.Method))
		ctx.Next()
		LoggerWithRequest(r).Info("Response", zap.Int("status", ctx.Writer.Status()))
	}
}
