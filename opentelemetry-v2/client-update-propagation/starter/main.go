package main

import (
	"context"
	"fmt"
	"log"
	"time"

	otelsetup "github.com/temporalio/samples-go/opentelemetry-v2"
	clientupdate "github.com/temporalio/samples-go/opentelemetry-v2/client-update-propagation"
	"go.opentelemetry.io/otel"
	"go.temporal.io/sdk/client"
	temporalotel "go.temporal.io/sdk/contrib/opentelemetry-v2"
)

const (
	instrumentationName = "github.com/temporalio/samples-go/opentelemetry-v2/client-update-propagation/starter"
	serviceName         = "temporal-otel-v2-custom-client-update-propagation-client"
	shutdownTimeout     = 5 * time.Second
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

	we, err := c.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{TaskQueue: clientupdate.TaskQueueName},
		clientupdate.Workflow,
	)
	if err != nil {
		return fmt.Errorf("unable to execute workflow: %w", err)
	}

	updateResult, err := sendUpdate(ctx, c, we.GetID(), "Temporal")
	if err != nil {
		return err
	}

	var workflowResult string
	if err := we.Get(ctx, &workflowResult); err != nil {
		return fmt.Errorf("unable to get workflow result: %w", err)
	}

	log.Println("Update result:", updateResult)
	log.Println("Workflow result:", workflowResult)
	return nil
}

// @@@SNIPSTART samples-go-opentelemetry-v2-client-update
func sendUpdate(
	ctx context.Context,
	c client.Client,
	workflowID string,
	name string,
) (string, error) {
	ctx, span := otel.Tracer(instrumentationName).Start(ctx, "send-update")
	defer span.End()

	handle, err := c.UpdateWorkflow(ctx, client.UpdateWorkflowOptions{
		WorkflowID:   workflowID,
		UpdateName:   clientupdate.UpdateName,
		WaitForStage: client.WorkflowUpdateStageCompleted,
		Args:         []interface{}{name},
	})
	if err != nil {
		return "", fmt.Errorf("unable to send Workflow Update: %w", err)
	}

	var result string
	if err := handle.Get(ctx, &result); err != nil {
		return "", fmt.Errorf("unable to get Workflow Update result: %w", err)
	}
	return result, nil
}

// @@@SNIPEND
