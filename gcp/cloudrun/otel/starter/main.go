package main

import (
	"context"
	"log"

	greeting "github.com/temporalio/samples-go/gcp/cloudrun/otel/greeting"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
)

// Helper program to start a workflow execution against the Cloud Run worker.
func main() {
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "cloud-run-workflow-id",
		TaskQueue: "cloud-run-task-queue",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, greeting.SampleWorkflow, "Cloud Run Worker!")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	var result string
	if err := we.Get(context.Background(), &result); err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
