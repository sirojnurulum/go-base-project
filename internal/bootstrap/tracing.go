package bootstrap

import (
	"io"

	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// InitTracerProvider initializes an OpenTelemetry TracerProvider with conditional output.
func InitTracerProvider(serviceName string, enableDetailedTracing bool) (*tracesdk.TracerProvider, error) {
	var exporter tracesdk.SpanExporter
	var err error

	if enableDetailedTracing {
		// Create a new exporter to send trace data to the console for development.
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, err
		}
		log.Println("Tracer provider initialized with detailed console output")
	} else {
		// Create a silent exporter that discards trace data for production/cleaner logs.
		exporter, err = stdouttrace.New(
			stdouttrace.WithWriter(io.Discard),
			stdouttrace.WithoutTimestamps(),
		)
		if err != nil {
			return nil, err
		}
		log.Println("Tracer provider initialized with silent mode")
	}

	// Create a new TracerProvider with a batch span processor and the configured exporter.
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)

	// Set the global TracerProvider.
	otel.SetTracerProvider(tp)

	return tp, nil
}
