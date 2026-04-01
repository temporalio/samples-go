package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	retryactivitynohb "github.com/temporalio/samples-go/retryactivity-no-heartbeat"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "retry-activity-no-heartbeat", worker.Options{})

	w.RegisterWorkflow(retryactivitynohb.RetryWorkflow)
	w.RegisterActivity(retryactivitynohb.BatchProcessingActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
