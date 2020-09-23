package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/splitmerge"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "split_merge_" + uuid.New(),
		TaskQueue: "split-merge",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, splitmerge.SampleSplitMergeWorkflow, 5)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
