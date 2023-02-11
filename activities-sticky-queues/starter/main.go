package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"

	activities_sticky_queues "github.com/temporalio/samples-go/activities-sticky-queues"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "activities_sticky_queues_WorkflowID",
		TaskQueue: "activities-sticky-queues",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, activities_sticky_queues.FileProcessingWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow completed")
}
