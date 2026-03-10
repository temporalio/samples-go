package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/temporalio/samples-go/searchattributes"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "search_attributes_" + uuid.New(),
		TaskQueue: "search-attributes",
		TypedSearchAttributes: temporal.NewSearchAttributes(
			searchattributes.CustomIntField.ValueSet(1),
		),
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
