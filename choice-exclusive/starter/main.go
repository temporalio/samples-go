package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

	choice "github.com/temporalio/temporal-go-samples/choice-exclusive"
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
		ID:        "exclusive_" + uuid.New(),
		TaskQueue: "choice",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, choice.ExclusiveChoiceWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	} else {
		log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	}

}
