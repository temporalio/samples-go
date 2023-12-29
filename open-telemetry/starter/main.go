package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"

	open_telemetry "github.com/temporalio/samples-go/open-telemetry"
)

func main() {
	ctx := context.Background()
	tracerProvider, err := open_telemetry.CreateTraceProvider(ctx)
	if err != nil {
		log.Fatalln("Unable to create trace provider", err)
	}
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Fatalln("Unable to shutdown trace provider", err)
		}
	}()

	tracer, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
		Tracer: tracerProvider.Tracer("go-sdk"),
	})
	if err != nil {
		log.Fatalln("Unable to create open telemetry interceptor", err)
	}
	// The client is a heavyweight object that should be created only once per process.
	c, err := client.Dial(client.Options{
		HostPort:     client.DefaultHostPort,
		Interceptors: []interceptor.ClientInterceptor{tracer},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// This Workflow ID can be a user supplied business logic identifier.
	workflowID := "open_telemetry_workflow_" + uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "open-telemetry",
	}

	workflowRun, err := c.ExecuteWorkflow(context.Background(), workflowOptions, open_telemetry.Workflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow",
		"WorkflowID", workflowRun.GetID(), "RunID", workflowRun.GetRunID())

	// Synchronously wait for the Workflow Execution to complete.
	// Behind the scenes the SDK performs a long poll operation.
	// If you need to wait for the Workflow Execution to complete from another process use
	// Client.GetWorkflow API to get an instance of the WorkflowRun.
	err = workflowRun.Get(context.Background(), nil)
	if err != nil {
		log.Fatalln("Failure getting workflow result", err)
	}
	log.Printf("Workflow finished")
}
