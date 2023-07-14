package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"

	otelworkflow "github.com/temporalio/samples-go/opentelemetry"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	otelworkflow.Setup(ctx)
	defer otelworkflow.Shutdown(ctx)

	options := client.Options{
		Interceptors: []interceptor.ClientInterceptor{otelworkflow.GetInterceptor()},
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
