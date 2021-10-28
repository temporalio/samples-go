package main

import (
	"log"

	"github.com/temporalio/samples-go/encryption"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		// Set DataConverter here so that workflow and activity inputs/results can
		// be encrypted/decrypted as required.
		DataConverter: encryption.NewEncryptionDataConverter(
			converter.GetDefaultDataConverter(),
			encryption.DataConverterOptions{Compress: true},
		),
		ContextPropagators: []workflow.ContextPropagator{encryption.NewContextPropagator()},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "encryption", worker.Options{})

	w.RegisterWorkflow(encryption.Workflow)
	w.RegisterActivity(encryption.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
