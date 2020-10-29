package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/metrics"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client.", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "metrics_workflowID",
		TaskQueue: "metrics",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, metrics.Workflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow.", err)
	}

	log.Println("Started workflow.", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	err = we.Get(context.Background(), nil)
	if err != nil {
		log.Fatalln("Unable to wait for workflow completition.", err)
	}

	log.Println("Check metrics at http://localhost:9090/metrics")
}
