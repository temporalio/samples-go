package main

import (
	"context"
	"log"

	"github.com/temporalio/samples-go/polling/frequent"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "FrequentPollingSampleWorkflow" + uuid.New(),
		TaskQueue: frequent.TaskQueueName,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, frequent.FrequentPolling)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
