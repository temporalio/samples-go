package main

import (
	"log"

	"github.com/temporalio/samples-go/schedule/overlap"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, overlap.TaskQueue, worker.Options{})
	w.RegisterWorkflow(overlap.PolicyWorkflow)

	log.Printf("Polling task queue %q. Press Ctrl+C to stop.", overlap.TaskQueue)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
