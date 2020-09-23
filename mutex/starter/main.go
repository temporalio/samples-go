package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/mutex"
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

	// This workflow ID can be user business logic identifier as well.
	resourceID := uuid.New()
	workflow1Options := client.StartWorkflowOptions{
		ID:        "SampleWorkflow1WithMutex_" + uuid.New(),
		TaskQueue: "mutex",
	}

	workflow2Options := client.StartWorkflowOptions{
		ID:        "SampleWorkflow2WithMutex_" + uuid.New(),
		TaskQueue: "mutex",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflow1Options, mutex.SampleWorkflowWithMutex, resourceID)
	if err != nil {
		log.Fatalln("Unable to execute workflow1", err)
	} else {
		log.Println("Started workflow1", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	}

	we, err = c.ExecuteWorkflow(context.Background(), workflow2Options, mutex.SampleWorkflowWithMutex, resourceID)
	if err != nil {
		log.Fatalln("Unable to execute workflow2", err)
	} else {
		log.Println("Started workflow2", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	}
}
