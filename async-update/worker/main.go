package main

import (
	"log"

	async_update "github.com/temporalio/samples-go/async-update"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "async-update", worker.Options{})

	w.RegisterWorkflow(async_update.ProcessWorkflow)
	w.RegisterActivity(async_update.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
