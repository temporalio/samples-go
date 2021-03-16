package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"

	cryptconverter "github.com/temporalio/samples-go/encrypted-payloads"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		// Set DataConverter here so that workflow and activity inputs/results can
		// be encrypted/decrypted as required.
		DataConverter: cryptconverter.NewCryptDataConverter(
			converter.GetDefaultDataConverter(),
		),
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "encrypted-payloads", worker.Options{})

	w.RegisterWorkflow(cryptconverter.Workflow)
	w.RegisterActivity(cryptconverter.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
