package main

import (
	"log"

	"github.com/temporalio/samples-go/snappycompress"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		// Set DataConverter here so that workflow and activity inputs/results will
		// be compressed as required.
		DataConverter: snappycompress.AlwaysCompressDataConverter,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "snappycompress", worker.Options{})

	w.RegisterWorkflow(snappycompress.Workflow)
	w.RegisterActivity(snappycompress.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
