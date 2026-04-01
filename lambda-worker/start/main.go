package main

import (
	"context"
	"fmt"
	"log"

	greeting "github.com/temporalio/samples-go/lambda-worker/greeting"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"
)

// This is a helper program to start a workflow execution
func main() {
	// Create a Temporal client
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create Temporal client", err)
	}
	defer c.Close()
	fmt.Printf("✅ Connected to Temporal Service\n")

	// Start the workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        "serverless-workflow-id-1",
		TaskQueue: "serverless-task-queue-1",
		VersioningOverride: &client.PinnedVersioningOverride{
			Version: worker.WorkerDeploymentVersion{
				DeploymentName: "my-app",
				BuildID:        "build-1",
			},
		},
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, greeting.SampleWorkflow, "Serverless Lambda Worker!")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Wait for workflow completion
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}

	log.Println("Workflow result:", result)
}
