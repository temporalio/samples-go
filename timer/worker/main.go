package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/timer"
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

	w := worker.New(c, "timer", worker.Options{
		MaxConcurrentActivityExecutionSize: 3,
	})

	w.RegisterWorkflow(timer.SampleTimerWorkflow)
	w.RegisterActivity(timer.OrderProcessingActivity)
	w.RegisterActivity(timer.SendEmailActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
