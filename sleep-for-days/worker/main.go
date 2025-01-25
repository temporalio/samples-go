package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	sleepfordays "github.com/temporalio/samples-go/sleep-for-days"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "sleep-for-days", worker.Options{})

	w.RegisterWorkflow(sleepfordays.SleepForDaysWorkflow)
	w.RegisterActivity(sleepfordays.SendEmailActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
