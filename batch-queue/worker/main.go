package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	batch_queue "github.com/temporalio/samples-go/batch-queue"
)

func main() {

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "batch", worker.Options{})

	w.RegisterWorkflow(batch_queue.AccumulateAndBatchWorkflow)
	w.RegisterWorkflow(batch_queue.SignalNewValuesWorkflow)
	w.RegisterActivity(batch_queue.WriteBatchToFile)
	w.RegisterActivity(batch_queue.WriteValToFile)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
