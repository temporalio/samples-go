package main

import (
	"log"

	remotecodec "github.com/temporalio/samples-go/remote-codec"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		// Set DataConverter here so that workflow and activity inputs/results will
		// be compressed as required.
		DataConverter: remotecodec.DataConverter,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "remotecodec", worker.Options{})

	w.RegisterWorkflow(remotecodec.Workflow)
	w.RegisterActivity(remotecodec.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
