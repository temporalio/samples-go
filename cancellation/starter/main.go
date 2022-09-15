package main

import (
	"context"
	"flag"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/cancellation"
)

// @@@SNIPSTART samples-go-cancellation-workflow-execution-starter
func main() {
	var workflowID string
	flag.StringVar(&workflowID, "w", "workflowID-to-cancel", "w is the workflowID of the workflow to be canceled.")
	flag.Parse()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "cancel-activity",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, cancellation.YourWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}

// @@@SNIPEND
