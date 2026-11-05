package main

import (
	"context"
	"flag"
	requestcancelexternalworkflow "github.com/temporalio/samples-go/request-cancel-external-workflow"
	"log"

	"go.temporal.io/sdk/client"
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

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, requestcancelexternalworkflow.CancellingWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}

// @@@SNIPEND
