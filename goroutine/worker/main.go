package main

import (
	"github.com/temporalio/samples-go/goroutine"
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

	w := worker.New(c, "goroutine", worker.Options{})

	w.RegisterWorkflow(goroutine.SampleGoroutineWorkflow)
	w.RegisterActivity(goroutine.Step1)
	w.RegisterActivity(goroutine.Step2)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
