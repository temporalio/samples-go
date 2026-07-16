package main

import (
	"context"
	"log"
	"time"

	"github.com/pborman/uuid"
	"github.com/temporalio/samples-go/helloworld"
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
		ID:        "startdelay_" + uuid.New(),
		TaskQueue: "startdelay",
		// The first workflow task will be dispatched in 5 minutes
		StartDelay: 5 * time.Minute,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, helloworld.Workflow, "from a delayed workflow")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	// ExecuteWorkflow will return immediately, but the workflow won't start executing till the StartDelay expires.
	log.Println("Scheduled workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
