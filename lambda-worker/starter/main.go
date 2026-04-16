package main

import (
	"context"
	"fmt"
	"log"
	"wci-test-function/v2/app"

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
	fmt.Printf("✅ Connected to Temporal Service")

	// Start the workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        "demo-run-1",
		TaskQueue: "tq-demo",
		VersioningOverride: &client.PinnedVersioningOverride{
			Version: worker.WorkerDeploymentVersion{
				DeploymentName: "demo-order",
				BuildID:        "1.0.0",
			},
		},
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, app.GreetingWorkflow, "Serverless Lambda Worker!")
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
