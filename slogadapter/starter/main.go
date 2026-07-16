package main

import (
	"context"
	"log"

	"github.com/temporalio/samples-go/slogadapter"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "slog_logger_workflow_id",
		TaskQueue: "slog-logger",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, slogadapter.Workflow, "<param to log>")
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
