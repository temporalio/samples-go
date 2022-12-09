package main

import (
	"context"
	"github.com/temporalio/samples-go/goroutine"
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
		ID:        "goroutine-" + uuid.New(),
		TaskQueue: "goroutine",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, goroutine.SampleGoroutineWorkflow, 5)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
