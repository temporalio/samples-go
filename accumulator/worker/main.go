package main

import (
	"log"

	accumulator "github.com/temporalio/samples-go/accumulator"

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

	w := worker.New(c, "accumulate_greetings", worker.Options{})

	w.RegisterWorkflow(accumulator.AccumulateSignalsWorkflow)
	w.RegisterActivity(accumulator.ComposeGreeting)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
