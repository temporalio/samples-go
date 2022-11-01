package main

import (
	"context"
	"fmt"
	"go.temporal.io/sdk/converter"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

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

	resp, err := c.DescribeWorkflowExecution(context.Background(), we.GetID(), we.GetRunID())
	if err != nil {
		log.Fatalln("Unable to desc workflow", err)
	}
	searchAttributes := resp.GetWorkflowExecutionInfo().GetSearchAttributes()

	for key, value := range searchAttributes.IndexedFields {
		var object interface{}
		err := converter.GetDefaultDataConverter().FromPayload(value, &object)
		if err != nil {
			log.Fatalln("Unable to convert to object", err)
		}

		str, isString := object.(string)
		if isString {
			fmt.Println("got a string, ", key, str)
		}
		f, isFloat := object.(float64)
		if isFloat {
			fmt.Println("got a float, ", key, f)
		}
	}
}
