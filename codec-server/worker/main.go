package main

import (
	"log"

	codecserver "github.com/temporalio/samples-go/codec-server"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		// Set DataConverter here so that workflow and activity inputs/results will
		// be compressed as required.
		DataConverter: codecserver.DataConverter,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "codecserver", worker.Options{})

	w.RegisterWorkflow(codecserver.Workflow)
	w.RegisterActivity(codecserver.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
