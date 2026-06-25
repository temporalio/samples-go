package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"

	batch_queue "github.com/temporalio/samples-go/batch-queue"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// start forever batching workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:        batch_queue.GetAccumulateAndBatchWorkflowID(),
		TaskQueue: "batch",
	}
	_, err = c.ExecuteWorkflow(context.Background(), workflowOptions, batch_queue.AccumulateAndBatchWorkflow, nil)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	// start signaling workflow
	workflowOptions = client.StartWorkflowOptions{
		ID:        batch_queue.GetSignalNewValuesWorkflowID(),
		TaskQueue: "batch",
	}
	_, err = c.ExecuteWorkflow(context.Background(), workflowOptions, batch_queue.SignalNewValuesWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
}
