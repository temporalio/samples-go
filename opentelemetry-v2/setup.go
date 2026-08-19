package opentelemetryv2

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.39.0"
	temporalotel "go.temporal.io/sdk/contrib/opentelemetry-v2"
)

// InitializeGlobalTracerProvider installs the replay-safe provider required by
// the OpenTelemetry v2 plugin and Workflow tracer.
// @@@SNIPSTART samples-go-opentelemetry-v2-tracer-provider
func InitializeGlobalTracerProvider(
	ctx context.Context,
	serviceName string,
) (*temporalotel.ReplaySafeTracerProvider, error) {
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("127.0.0.1:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	provider := temporalotel.NewReplaySafeTracerProvider(
		// WithBatcher performs exporter I/O outside the Workflow goroutine.
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)
	otel.SetTracerProvider(provider)
	return provider, nil
}

// @@@SNIPEND

// InitializeGlobalMeterProvider installs a Prometheus-backed meter provider.
// @@@SNIPSTART samples-go-opentelemetry-v2-meter-provider
func InitializeGlobalMeterProvider(serviceName string) (*sdkmetric.MeterProvider, error) {
	exporter, err := otelprometheus.New()
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(exporter),
		sdkmetric.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)
	otel.SetMeterProvider(provider)
	return provider, nil
}

// @@@SNIPEND

// StartPrometheusEndpoint exposes registered metrics in Prometheus format.
// @@@SNIPSTART samples-go-opentelemetry-v2-metrics-server
func StartPrometheusEndpoint() (*http.Server, error) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:              "127.0.0.1:9090",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return nil, err
	}

	log.Println("Prometheus metrics available at http://127.0.0.1:9090/metrics")
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println("Prometheus endpoint failed:", err)
		}
	}()
	return server, nil
}

// @@@SNIPEND
