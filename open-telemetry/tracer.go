package open_telemetry

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func CreateTraceProvider(ctx context.Context) (*trace.TracerProvider, error) {
	exp, err := otlptracegrpc.New(ctx, otlptracegrpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return trace.NewTracerProvider(
		trace.WithBatcher(exp, trace.WithBatchTimeout(time.Second)),
		trace.WithResource(resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("temporal-sample"),
		))), nil
}
