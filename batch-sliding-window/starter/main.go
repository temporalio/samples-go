package main

import (
	"context"
	"log"

	batch_sliding_window "github.com/temporalio/samples-go/batch-sliding-window"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created only once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "batch-sliding-window",
	}
	ctx := context.Background()
	input := batch_sliding_window.ProcessBatchWorkflowInput{
		PageSize:          3,
		SlidingWindowSize: 2,
		Partitions:        1,
	}
	we, err := c.ExecuteWorkflow(ctx, workflowOptions, batch_sliding_window.ProcessBatchWorkflow, input)
	if err != nil {
		log.Fatalln("Failure starting workflow", err)
	}
	log.Println("Started Workflow Execution", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Wait for Workflow Execution completion.
	// This is rarely needed in real use cases as batch workflows are usually long-running.
	var result int
	err = we.Get(ctx, &result)
	if err != nil {
		panic(err)
	}
	log.Println("Completed workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID(), "Result", result)
}
