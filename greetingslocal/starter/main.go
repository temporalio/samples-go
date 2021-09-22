package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"github.com/temporalio/samples-go/greetingslocal"
	"go.temporal.io/sdk/client"
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

	workflowOptions := client.StartWorkflowOptions{
		ID:        "greetings_" + uuid.New(),
		TaskQueue: "greetings",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, greetingslocal.GreetingSample)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
