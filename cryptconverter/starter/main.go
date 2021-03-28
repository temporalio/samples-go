package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"

	"github.com/temporalio/samples-go/cryptconverter"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		// Set DataConverter here to ensure that workflow inputs and results are
		// encrypted/decrypted as required.
		DataConverter: cryptconverter.NewCryptDataConverter(
			converter.GetDefaultDataConverter(),
		),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "cryptconverter_workflowID",
		TaskQueue: "cryptconverter",
	}

	// The workflow input "My Secret Friend" will be encrypted by the DataConverter before being sent to Temporal
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, cryptconverter.Workflow, "My Secret Friend")
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
