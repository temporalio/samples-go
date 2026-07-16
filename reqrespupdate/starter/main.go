package main

import (
	"context"
	"log"

	"github.com/temporalio/samples-go/reqrespupdate"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "reqrespupdate_workflow",
		TaskQueue: "reqrespupdate",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, reqrespupdate.UppercaseWorkflow, true)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
