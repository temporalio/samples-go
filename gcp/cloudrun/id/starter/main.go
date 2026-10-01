package main

import (
	"context"
	"log"
	"os"

	greeting "github.com/temporalio/samples-go/gcp/cloudrun/id/greeting"

	"go.temporal.io/sdk/client"
)

// Helper program to start a workflow execution against the Cloud Run worker.
func main() {
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

	var result string
	if err := we.Get(context.Background(), &result); err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}
	log.Println("Workflow result:", result)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
