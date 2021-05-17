package main

import (
	"log"

	largeeventhistory "github.com/temporalio/samples-go/temporal-fixtures/large-event-history"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "largepayload", worker.Options{})

	w.RegisterWorkflow(largeeventhistory.Workflow)
	w.RegisterActivity(&largeeventhistory.Activities{})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
