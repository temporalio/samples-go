package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	typedsearchattributes "github.com/temporalio/samples-go/typed-searchattributes"
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
		ID:                    "typed-search_attributes_" + uuid.New(),
		TaskQueue:             "typed-search-attributes",
		TypedSearchAttributes: temporal.NewSearchAttributes(typedsearchattributes.CustomIntKey.ValueSet(1)),
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, typedsearchattributes.SearchAttributesWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
