package main

import (
	"context"
	"log"

	otelworkflow "github.com/temporalio/samples-go/opentelemetry"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentelemetry"
	"go.temporal.io/sdk/interceptor"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := otelworkflow.InitializeGlobalTracerProvider()
	if err != nil {
		log.Fatalln("Unable to create a global trace provider", err)
	}
	defer tp.Shutdown(ctx)

	tracingInterceptor, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{})
	if err != nil {
		log.Fatalln("Unable to create interceptor", err)
	}

	options := client.Options{
		Interceptors: []interceptor.ClientInterceptor{tracingInterceptor},
	}

	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(options)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "otel_workflowID",
		TaskQueue: "otel",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, otelworkflow.Workflow, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
