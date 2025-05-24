package main

import (
	"context"
	securityinterceptor "github.com/temporalio/samples-go/logger-interceptor"
	"log"

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
		TaskQueue: "security-interceptor",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, securityinterceptor.Workflow, "Temporal")
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
	log.Println("Workflow result:", result)
}
