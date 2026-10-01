// @@@SNIPSTART go-cloud-run-worker
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	greeting "github.com/temporalio/samples-go/gcp/cloudrun/otel/greeting"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/contrib/gcp/cloudrun/otel"
	"go.temporal.io/sdk/worker"
)

const taskQueue = "cloud-run-task-queue"

func main() {
	ctx := context.Background()

	// Exports OTLP metrics and traces to the collector sidecar (defaults to localhost:4317).
	otelPlugin, err := otel.NewPlugin(ctx, otel.PluginOptions{})
	if err != nil {
		log.Fatalln("Unable to create OpenTelemetry plugin", err)
	}

	// A client plugin that also implements worker.Plugin is applied to workers automatically.
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

	// Reserve most of the termination window to flush telemetry before the process is killed.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := otelPlugin.Shutdown(shutdownCtx); err != nil {
		log.Println("Failed to shut down OpenTelemetry plugin:", err)
	}
	log.Println("Worker stopped")
}

// @@@SNIPEND
