// @@@SNIPSTART go-cloud-run-worker
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	greeting "github.com/temporalio/samples-go/cloud-run-worker/greeting"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/contrib/gcp/cloudrun"
	"go.temporal.io/sdk/worker"
)

const taskQueue = "cloud-run-task-queue"

func main() {
	ctx := context.Background()

	// The plugin exports OTLP metrics and traces to the collector sidecar on
	// localhost:4317. The endpoint and service name fall back to environment
	// variables (OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_SERVICE_NAME,
	// CLOUD_RUN_WORKER_POOL, K_SERVICE) that Cloud Run provides.
	otelPlugin, err := cloudrun.NewOpenTelemetryPlugin(ctx, cloudrun.OpenTelemetryPluginOptions{})
	if err != nil {
		log.Fatalln("Unable to create OpenTelemetry plugin", err)
	}

	// Load the Temporal connection from the environment (see temporal.toml or the
	// TEMPORAL_* environment variables) and install the plugin. Client plugins
	// that also implement worker.Plugin are applied to workers automatically.
	clientOptions, err := envconfig.LoadDefaultClientOptions()
	if err != nil {
		log.Fatalln("Unable to load Temporal client options", err)
	}
	clientOptions.Plugins = append(clientOptions.Plugins, otelPlugin)

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}

	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(greeting.SampleWorkflow)
	w.RegisterActivity(greeting.HelloActivity)

	if err := w.Start(); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
	log.Println("Worker started on task queue", taskQueue)

	// Cloud Run sends SIGTERM and allows roughly ten seconds before SIGKILL.
	signalCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-signalCtx.Done()

	log.Println("Shutdown signal received, stopping worker")
	w.Stop()
	c.Close()

	// Reserve most of Cloud Run's termination window to flush telemetry before
	// the process is killed.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := otelPlugin.Shutdown(shutdownCtx); err != nil {
		log.Println("Failed to shut down OpenTelemetry plugin:", err)
	}
	log.Println("Worker stopped")
}

// @@@SNIPEND
