package main

import (
	"log"

	rainbowstatuses "github.com/temporalio/samples-go/temporal-fixtures/rainbow-statuses"
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

	w := worker.New(c, "rainbow-statuses", worker.Options{})

	w.RegisterWorkflow(rainbowstatuses.RainbowStatusesWorkflow)
	w.RegisterActivity(&rainbowstatuses.Activities{})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
