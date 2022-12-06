package main

import (
	"context"
	"log"

	"github.com/temporalio/samples-go/helloworld"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
	ctx := context.Background()

	numberOfWorkflows := 10
	workflowOptions := client.StartWorkflowOptions{
		ID:        "multiple_history_replay_workflowID",
		TaskQueue: "multiple-history-replay",
	}

	// Simulate existing workflows on the server
	for w := 0; w < numberOfWorkflows; w++ {
		we, err := c.ExecuteWorkflow(ctx, workflowOptions, helloworld.Workflow, "Temporal")
		if err != nil {
			log.Fatalln("Unable to execute workflow", err)
		}
		log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

		// Wait for the workflow to finish
		err = we.Get(ctx, nil)
		if err != nil {
			log.Fatalln("Unable get workflow result", err)
		}

		log.Println("Workflow finished", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	}
}
