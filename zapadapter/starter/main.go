package main

import (
	"context"
	"log"

	"github.com/temporalio/samples-go/zapadapter"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "zap_logger_workflow_id",
		TaskQueue: "zap-logger",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, zapadapter.Workflow, "<param to log>")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result interface{}
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow completed. Check worker logs.")
}
