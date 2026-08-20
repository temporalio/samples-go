package main

import (
	"context"
	"fmt"
	"log"
	"time"

	otelsetup "github.com/temporalio/samples-go/opentelemetry-v2"
	clientupdate "github.com/temporalio/samples-go/opentelemetry-v2/client-update-propagation"
	"go.temporal.io/sdk/client"
	temporalotel "go.temporal.io/sdk/contrib/opentelemetry-v2"
	"go.temporal.io/sdk/worker"
)

const (
	serviceName     = "temporal-otel-v2-custom-client-update-propagation-worker"
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

	plugin, err := temporalotel.NewPlugin(temporalotel.PluginOptions{})
	if err != nil {
		return fmt.Errorf("unable to create plugin: %w", err)
	}
	c, err := client.Dial(client.Options{Plugins: []client.Plugin{plugin}})
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}
	defer c.Close()

	w := worker.New(c, clientupdate.TaskQueueName, worker.Options{})
	w.RegisterWorkflow(clientupdate.Workflow)

	if err := w.Run(worker.InterruptCh()); err != nil {
		return fmt.Errorf("worker run failed: %w", err)
	}
	return nil
}
