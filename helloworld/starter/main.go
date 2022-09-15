package main

import (
	"context"
	"errors"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/temporalio/samples-go/helloworld"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello_world_workflowID",
		TaskQueue: "hello-world",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, helloworld.Workflow, "Temporal")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Println("Unable get workflow result", err)

		var nonRetryErr *helloworld.NonRetryableError
		if errors.As(err, &nonRetryErr) {
			log.Println("NonRetryableError", nonRetryErr.Err, nonRetryErr.Code)
		} else {
			log.Println("Cannot convert to NonRetryableError")
		}

		var appErr *temporal.ApplicationError
		if errors.As(err, &appErr) {
			log.Println("ApplicationError", appErr)
			log.Println("HasDetails", appErr.HasDetails())
			err = appErr.Details(&nonRetryErr)
			log.Println("conversion err?", err)
			log.Println("NonRetryableError", nonRetryErr.Err, *nonRetryErr.Code)
		} else {
			log.Println("Cannot convert to ApplicationError")
		}
	}
	log.Println("Workflow result:", result)
}
