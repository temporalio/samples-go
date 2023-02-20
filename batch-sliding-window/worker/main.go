package main

import (
	batch_sliding_window "github.com/temporalio/samples-go/batch-sliding-window"
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

	w := worker.New(c, "batch-sliding-window", worker.Options{})

	w.RegisterWorkflow(batch_sliding_window.ProcessBatchWorkflow)
	w.RegisterWorkflow(batch_sliding_window.SlidingWindowWorkflow)
	w.RegisterWorkflow(batch_sliding_window.RecordProcessorWorkflow)

	w.RegisterActivity(&batch_sliding_window.RecordLoader{RecordCount: 10})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
