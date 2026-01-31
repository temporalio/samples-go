package main

import (
	"context"
	"log"
	"os"

	dataconverterlargepayloads "github.com/temporalio/samples-go/data-converter-large-payloads"
	pc "github.com/temporalio/samples-go/data-converter-large-payloads/payloadconverter"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
)

func main() {

	dataconverter := converter.NewCompositeDataConverter(
		converter.NewNilPayloadConverter(),
		converter.NewByteSlicePayloadConverter(),
		pc.NewLargeSizePayloadConverter(),
		// fallback converter for payloads that do not exceed the threshold size
		converter.NewJSONPayloadConverter(),
	)

	c, err := client.Dial(client.Options{
		HostPort:      client.DefaultHostPort,
		DataConverter: dataconverter,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowID := "default_workflow_id"
	if len(os.Args) > 1 {
		workflowID = os.Args[1]
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "data-converter-large-payloads",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, dataconverterlargepayloads.Workflow)
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
}
