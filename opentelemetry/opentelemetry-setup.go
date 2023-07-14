package opentelemetry

import (
	"context"
	"log"

	"go.temporal.io/sdk/contrib/opentelemetry"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.temporal.io/sdk/interceptor"
)

var tracingInterceptor interceptor.ClientInterceptor

func Setup(ctx context.Context) {
	// Initialize tracer
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		log.Fatalln("failed to initialize stdouttrace exporter", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("temporal-example"),
			semconv.ServiceVersion("0.0.1"),
		)),
	)
	otel.SetTracerProvider(tp)

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	tracingInterceptor, err = opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
	if err != nil {
		log.Fatalln("Unable to create interceptor", err)
	}
}

func GetInterceptor() interceptor.ClientInterceptor {
	return tracingInterceptor
}

func Shutdown(ctx context.Context) {
	tp := otel.GetTracerProvider().(*sdktrace.TracerProvider)
	_ = tp.Shutdown(ctx)
}
