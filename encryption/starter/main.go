package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/encryption"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		// If you intend to use the same encryption key for all workflows you can
		// set the KeyID for the encryption encoder like so:
		//
		// Set DataConverter to ensure that workflow inputs and results are
		// encrypted/decrypted as required.
		//
		//   DataConverter: encryption.NewEncryptionDataConverter(
		// 	  converter.GetDefaultDataConverter(),
		// 	  encryption.DataConverterOptions{KeyID: "test", Compress: true},
		//   ),
		//
		// In this case you do not need to use a ContextPropagator.
		//
		// If you need to vary the encryption key per workflow, you can instead
		// leave the KeyID unset for the encoder and supply it via the workflow
		// context as shown below. For this use case you will also need to use a
		// ContextPropagator so that KeyID is also available in the context for activities.
		//
		// Set DataConverter to ensure that workflow inputs and results are
		// encrypted/decrypted as required.
		DataConverter: encryption.NewEncryptionDataConverter(
			converter.GetDefaultDataConverter(),
			encryption.DataConverterOptions{Compress: true},
		),
		// Use a ContextPropagator so that the KeyID value set in the workflow context is
		// also availble in the context for activities.
		ContextPropagators: []workflow.ContextPropagator{encryption.NewContextPropagator()},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "encryption_workflowID",
		TaskQueue: "encryption",
	}

	ctx := context.Background()
	// If you are using a ContextPropagator and varying keys per workflow you need to set
	// the KeyID to use for this workflow in the context:
	ctx = context.WithValue(ctx, encryption.PropagateKey, encryption.CryptContext{KeyID: "test"})

	// The workflow input "My Secret Friend" will be encrypted by the DataConverter before being sent to Temporal
	we, err := c.ExecuteWorkflow(
		ctx,
		workflowOptions,
		encryption.Workflow,
		"My Secret Friend",
	)
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
