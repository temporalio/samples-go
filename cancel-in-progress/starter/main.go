package main

import (
	"context"
	"go.temporal.io/sdk/client"
	"log"
	"math/rand"
	"strconv"
	"time"

	cancel_in_progress "github.com/temporalio/samples-go/cancel-in-progress"
)

func main() {
	// The client is a heavyweight object that should be created only once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// This Workflow ID must be deterministic because we don't want to start a new workflow every time.
	projectID := "my-project"

	workflowID := "parent-workflow_" + projectID
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "child-workflow",
	}

	// Start the workflow execution or send a signal to an existing workflow execution.
	workflowRun, err := c.SignalWithStartWorkflow(
		context.Background(),
		workflowID,
		cancel_in_progress.ParentWorkflowSignalName,
		"World",
		workflowOptions,
		cancel_in_progress.SampleParentWorkflow,
	)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow",
		"WorkflowID", workflowRun.GetID(), "RunID", workflowRun.GetRunID())

	// Send three signals to the workflow. At the end, we will expect that only the result of the last signal is returned.
	for i := 1; i <= 3; i++ {
		workflowRun, err = c.SignalWithStartWorkflow(
			context.Background(),
			workflowID,
			cancel_in_progress.ParentWorkflowSignalName,
			"World"+strconv.Itoa(i),
			workflowOptions,
			cancel_in_progress.SampleParentWorkflow,
		)
		if err != nil {
			log.Fatalln("Unable to execute workflow", err)
		}

		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		log.Println("Started workflow",
			"WorkflowID", workflowRun.GetID(), "RunID", workflowRun.GetRunID())
	}

	// Synchronously wait for the Workflow Execution to complete.
	// Behind the scenes the SDK performs a long poll operation.
	// If you need to wait for the Workflow Execution to complete from another process use
	// Client.GetWorkflow API to get an instance of the WorkflowRun.
	var result string
	err = workflowRun.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Failure getting workflow result", err)
	}
	log.Printf("Workflow result: %v", result)
}
