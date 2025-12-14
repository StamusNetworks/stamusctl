// Package logging provides utilities for structured logging with OpenTelemetry integration.
package logging

import (
	"context"
	"net/http"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// LoggerWithRequest creates a logger with request context and tracing information.
func LoggerWithRequest(request *http.Request) otelzap.LoggerWithCtx {
	span := trace.SpanContextFromContext(request.Context())
	logger := Logger.With(
		zap.String("traceId", span.TraceID().String()),
		zap.String("spanId", span.SpanID().String()),
		zap.String("requestUri", request.RequestURI),
		zap.String("method", request.Method),
		zap.String("remoteAddr", request.RemoteAddr),
		zap.String("url", request.URL.String()),
		zap.String("referer", request.Header.Get("Referer")),
		zap.String("user-agent", request.Header.Get("User-Agent")),
		zap.String("x-request-id", request.Header.Get("X-Request-ID")),
	)

	return otelzap.New(logger).Ctx(request.Context())
}

// LoggerWithSpanContext creates a logger with trace and span ID from the given span context.
func LoggerWithSpanContext(span trace.SpanContext) *zap.Logger {
	return Logger.With(
		zap.String("traceId", span.TraceID().String()),
		zap.String("spanId", span.SpanID().String()),
	)
}

// LoggerWithContextToSpanContext creates a logger with trace and span ID from the given context.
func LoggerWithContextToSpanContext(ctx context.Context) *zap.Logger {
	span := trace.SpanContextFromContext(ctx)

	return Logger.With(
		zap.String("traceId", span.TraceID().String()),
		zap.String("spanId", span.SpanID().String()),
	)
}
