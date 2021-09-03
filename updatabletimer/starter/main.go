package main

import (
	"context"
	"github.com/temporalio/samples-go/updatabletimer"
	"log"
	"time"

	"go.temporal.io/sdk/client"
)

// Starts updatable timer workflow with initial wake-up time in 30 seconds.
func main() {
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        updatabletimer.WorkflowID,
		TaskQueue: updatabletimer.TaskQueue,
	}

	wakeUpTime := time.Now().Add(30 * time.Second)
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, updatabletimer.Workflow, wakeUpTime)
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	log.Println("Started workflow that is going to block on an updatable timer",
		"WorkflowID", we.GetID(), "RunID", we.GetRunID(), "WakeUpTime", wakeUpTime)
}
