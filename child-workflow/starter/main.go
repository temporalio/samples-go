package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

	child_workflow "github.com/temporalio/samples-go/child-workflow"
)

// @@@SNIPSTART samples-go-child-workflow-example-execution-starter
func main() {
	// The client is a heavyweight object that should be created only once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// This Workflow ID can be a user supplied business logic identifier.
	workflowID := "parent-workflow_" + uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "child-workflow",
	}

	workflowRun, err := c.ExecuteWorkflow(context.Background(), workflowOptions, child_workflow.SampleParentWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow",
		"WorkflowID", workflowRun.GetID(), "RunID", workflowRun.GetRunID())

	// Synchronously wait for the Workflow Execution to complete.
	// Behind the scenes the SDK performs a long poll operation.
	// If you need to wait for the Workflow Execution to complete from another process use
	// Client.GetWorkflow API to get an instance of the WorkflowRun.
	var result string
	err = workflowRun.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Failure getting workflow result", err)
	}
	log.Println("Workflow result: %v", "result", result)
}

// @@@SNIPEND
