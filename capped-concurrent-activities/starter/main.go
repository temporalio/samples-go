package main

import (
	"context"
	capped_activities "github.com/temporalio/samples-go/capped-concurrent-activities"
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
		ID:        "capped_activities_" + uuid.New(),
		TaskQueue: "capped-activities",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions,
		capped_activities.CappedActivitiesWorkflow, 90)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
