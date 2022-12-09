package main

import (
	splitmerge_selector "github.com/temporalio/samples-go/splitmerge-selector"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "split-merge-selector", worker.Options{})

	w.RegisterWorkflow(splitmerge_selector.SampleSplitMergeSelectorWorkflow)
	w.RegisterActivity(splitmerge_selector.ChunkProcessingActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
