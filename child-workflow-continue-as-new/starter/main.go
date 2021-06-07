package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

	cw "github.com/temporalio/samples-go/child-workflow-continue-as-new"
)
// @@@SNIPSTART samples-go-cw-cas-workflow-execution-starter
func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// This workflow ID can be user business logic identifier as well.
	workflowID := "parent-workflow_" + uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "child-workflow-continue-as-new",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, cw.SampleParentWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
// @@@SNIPEND
