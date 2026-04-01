package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

	retryactivitynohb "github.com/temporalio/samples-go/retryactivity-no-heartbeat"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "retry_activity_no_heartbeat_" + uuid.New(),
		TaskQueue: "retry-activity-no-heartbeat",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, retryactivitynohb.RetryWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
