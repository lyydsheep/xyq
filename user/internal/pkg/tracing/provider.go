package tracing

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// NewProvider creates a new OpenTelemetry trace provider
func NewProvider(endpoint, serviceName string, sampler float64) (*sdktrace.TracerProvider, error) {
	// Create Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(endpoint)))
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create tracer provider with batch span processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(sampler)),
	)

	// Register as global tracer provider
	otel.SetTracerProvider(tp)

	return tp, nil
}

// Shutdown properly shuts down the tracer provider
func Shutdown(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return tp.Shutdown(ctx)
}