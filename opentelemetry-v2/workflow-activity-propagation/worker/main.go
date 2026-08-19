package main

import (
	"context"
	"fmt"
	"log"
	"time"

	roototel "github.com/temporalio/samples-go/opentelemetry-v2"
	workflowactivity "github.com/temporalio/samples-go/opentelemetry-v2/workflow-activity-propagation"
	"go.temporal.io/sdk/client"
	temporalotel "go.temporal.io/sdk/contrib/opentelemetry-v2"
	"go.temporal.io/sdk/worker"
)

const (
	serviceName     = "temporal-otel-v2-custom-workflow-activity-propagation-worker"
	shutdownTimeout = 5 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	ctx := context.Background()
	tp, err := roototel.InitializeGlobalTracerProvider(ctx, serviceName)
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

	// @@@SNIPSTART samples-go-opentelemetry-v2-plugin-client
	plugin, err := temporalotel.NewPlugin(temporalotel.PluginOptions{})
	if err != nil {
		return fmt.Errorf("unable to create plugin: %w", err)
	}

	c, err := client.Dial(client.Options{Plugins: []client.Plugin{plugin}})
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}
	defer c.Close()
	// @@@SNIPEND

	w := worker.New(c, workflowactivity.TaskQueueName, worker.Options{})
	w.RegisterWorkflow(workflowactivity.Workflow)
	w.RegisterActivity(workflowactivity.Activity)

	if err := w.Run(worker.InterruptCh()); err != nil {
		return fmt.Errorf("worker run failed: %w", err)
	}
	return nil
}
