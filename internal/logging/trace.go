package logging

import (
	"runtime/debug"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	Tracer         trace.Tracer
	TracerProvider trace.TracerProvider
)

func NewTraceProvider() *tracesdk.TracerProvider {
	exporter, err := stdouttrace.New()
	if err != nil {
		Sugar.Error(err)
		debug.PrintStack()
		panic(err)
	}
	provider := tracesdk.WithBatcher(exporter)
	traceProvider := tracesdk.NewTracerProvider(provider)

	TracerProvider = traceProvider
	Tracer = traceProvider.Tracer("stamus-daemon")

	return traceProvider
}
