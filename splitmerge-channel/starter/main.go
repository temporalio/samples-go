package main

import (
	"context"
	splitmerge_channel "github.com/temporalio/samples-go/splitmerge-channel"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "split_merge_selector_" + uuid.New(),
		TaskQueue: "split-merge-selector",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, splitmerge_channel.SampleSplitMergeChannelWorkflow, 10)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	_ = we.Get(context.Background(), nil)
	log.Println(" Workflow Completed", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

}
