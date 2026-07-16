package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

	sessionfailure "github.com/temporalio/samples-go/session-failure"
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
		ID:        "session_failure_" + uuid.New(),
		TaskQueue: "session-failure",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, sessionfailure.SampleSessionFailureRecoveryWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
