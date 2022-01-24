package main

import (
	"context"
	"log"

	"github.com/temporalio/samples-go/reqrespquery"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "reqrespquery_workflow",
		TaskQueue: "reqrespquery",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, reqrespquery.UppercaseWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
