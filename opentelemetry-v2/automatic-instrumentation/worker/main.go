package main

import (
	"context"
	"fmt"
	"log"
	"time"

	otelsetup "github.com/temporalio/samples-go/opentelemetry-v2"
	automatic "github.com/temporalio/samples-go/opentelemetry-v2/automatic-instrumentation"
	"go.temporal.io/sdk/client"
	temporalotel "go.temporal.io/sdk/contrib/opentelemetry-v2"
	"go.temporal.io/sdk/interceptor/tracing"
	"go.temporal.io/sdk/worker"
)

const (
	serviceName     = "temporal-otel-v2-automatic-worker"
	shutdownTimeout = 5 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	ctx := context.Background()

	tp, err := otelsetup.InitializeGlobalTracerProvider(ctx, serviceName)
	if err != nil {
		return fmt.Errorf("unable to create global tracer provider: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := tp.Shutdown(shutdownCtx); err != nil {
			log.Println("Error shutting down tracer provider:", err)
		}
	}()

	mp, err := otelsetup.InitializeGlobalMeterProvider(serviceName)
	if err != nil {
		return fmt.Errorf("unable to create global meter provider: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := mp.Shutdown(shutdownCtx); err != nil {
			log.Println("Error shutting down meter provider:", err)
		}
	}()

	metricsEndpoint, err := otelsetup.StartPrometheusEndpoint()
	if err != nil {
		return fmt.Errorf("unable to start Prometheus endpoint: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := metricsEndpoint.Shutdown(shutdownCtx); err != nil {
			log.Println("Error shutting down Prometheus endpoint:", err)
		}
	}()

	// @@@SNIPSTART samples-go-opentelemetry-v2-metrics-plugin
	plugin, err := temporalotel.NewPlugin(temporalotel.PluginOptions{
		TracerOptions: tracing.TracerOptions{
			AddTemporalSpans: true,
		},
		MetricsHandlerOptions: &temporalotel.MetricsHandlerOptions{
			UseMonotonicCounters: true,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create plugin: %w", err)
	}
	// @@@SNIPEND

	c, err := client.Dial(client.Options{Plugins: []client.Plugin{plugin}})
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}
	defer c.Close()

	w := worker.New(c, automatic.TaskQueueName, worker.Options{})
	w.RegisterWorkflow(automatic.Workflow)
	w.RegisterActivity(automatic.Activity)

	if err := w.Run(worker.InterruptCh()); err != nil {
		return fmt.Errorf("worker run failed: %w", err)
	}
	return nil
}
