package main

import (
	"context"
	"log"

	batch_sliding_window "github.com/temporalio/samples-go/batch-sliding-window"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
)

func main() {
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		panic(err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "batch-sliding-window",
	}
	ctx := context.Background()
	input := batch_sliding_window.ProcessBatchWorkflowInput{
		PageSize:          5,
		SlidingWindowSize: 10,
		Partitions:        3,
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
