package logging

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Tracer is the global OpenTelemetry tracer instance.
var Tracer trace.Tracer
// TracerProvider is the global OpenTelemetry tracer provider instance.
var TracerProvider trace.TracerProvider

//nolint:ireturn // Interface return needed for OpenTelemetry span exporter
func createExporter(collectorURL string) tracesdk.SpanExporter {
	if collectorURL == "" {
		exporter, err := stdouttrace.New()
		if err != nil {
			Sugar.Fatal(err)
		}

		return exporter
	}
	secureOption := otlptracegrpc.WithInsecure()

	exporter, err := otlptrace.New(
		context.Background(),
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(collectorURL),
		),
	)
	if err != nil {
		Sugar.Fatal(err)
	}

	return exporter
}

// InitTracer initializes OpenTelemetry tracing with OTLP or stdout exporter.
func InitTracer(collectorURL, serviceName string) func(context.Context) error {
	exporter := createExporter(collectorURL)

	resources, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("library.language", "go"),
		),
	)
	if err != nil {
		Sugar.Fatal(err)
	}

	provider := tracesdk.WithBatcher(exporter)
	traceProvider := tracesdk.NewTracerProvider(tracesdk.WithSampler(tracesdk.AlwaysSample()),
		provider,
		tracesdk.WithResource(resources),
	)

	TracerProvider = traceProvider
	Tracer = traceProvider.Tracer("stamus-ctl")

	otel.SetTracerProvider(traceProvider)

	return exporter.Shutdown
}
