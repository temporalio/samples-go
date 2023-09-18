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

	fileID := uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:        "session_failure_" + fileID,
		TaskQueue: "session-failure",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, sessionfailure.SampleSessionFailureRecoveryWorkflow, fileID)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
