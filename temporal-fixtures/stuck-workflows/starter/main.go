package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

	stuckworkflows "github.com/temporalio/samples-go/temporal-fixtures/stuck-workflows"
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

	id := uuid.New()[:6]

	workflowOptions := client.StartWorkflowOptions{
		ID:        id + "_" + "stuck_activity",
		TaskQueue: "stuck-workflows",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, stuckworkflows.StuckWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	} else {
		log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	}

	workflowOptions = client.StartWorkflowOptions{
		ID:        id + "_" + "no_worker",
		TaskQueue: "no-worker",
	}

	we, err = c.ExecuteWorkflow(context.Background(), workflowOptions, stuckworkflows.StuckWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	} else {
		log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	}
}
