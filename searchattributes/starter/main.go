package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/searchattributes"
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

	workflowOptions := client.StartWorkflowOptions{
		ID:        "search_attributes_" + uuid.New(),
		TaskQueue: "search-attributes",
		SearchAttributes: map[string]interface{}{ // optional search attributes when start workflow
			"CustomIntField": 1,
		},
		Memo: map[string]interface{}{
			"description": "Test search attributes workflow",
		},
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, searchattributes.SearchAttributesWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
