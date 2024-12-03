package main

import (
	bsdc "github.com/temporalio/samples-go/blob-store-data-converter"
	"github.com/temporalio/samples-go/blob-store-data-converter/blobstore"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"log"
)

func main() {
	bsClient := blobstore.NewClient()

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		// Calls to the blob store will probably be a network call with inherent latency, this may trigger deadlock detection
		DataConverter: workflow.DataConverterWithoutDeadlockDetection(bsdc.NewDataConverter(
			converter.GetDefaultDataConverter(),
			bsClient,
		)),

		// Use a ContextPropagator so that the KeyID value set in the workflow context is
		// also available in the context for activities.
		ContextPropagators: []workflow.ContextPropagator{
			bsdc.NewContextPropagator(),
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "blobstore_codec", worker.Options{})

	w.RegisterWorkflow(bsdc.Workflow)
	w.RegisterActivity(bsdc.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
