package main

import (
	"context"
	"log"
	"os"

	greeting "github.com/temporalio/samples-go/cloud-run-worker-id/greeting"

	"go.temporal.io/sdk/client"
)

// This is a helper program to start a workflow execution against the Cloud Run worker.
func main() {
	// Connect using the same TEMPORAL_* environment variables the worker uses. This sample uses a
	// plaintext connection for simplicity.
	c, err := client.Dial(client.Options{
		HostPort:  getenv("TEMPORAL_ADDRESS", client.DefaultHostPort),
		Namespace: getenv("TEMPORAL_NAMESPACE", client.DefaultNamespace),
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "cloud-run-workflow-id",
		TaskQueue: getenv("TEMPORAL_TASK_QUEUE", "cloud-run-task-queue"),
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, greeting.SampleWorkflow, "Cloud Run Worker!")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Wait for workflow completion
	var result string
	if err := we.Get(context.Background(), &result); err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}
	log.Println("Workflow result:", result)
}

// getenv returns the value of the environment variable named by key, or fallback if it is unset or
// empty.
func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
