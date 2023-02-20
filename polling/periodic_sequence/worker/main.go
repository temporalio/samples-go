package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/polling"
	"github.com/temporalio/samples-go/polling/periodic_sequence"
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

	w := worker.New(c, periodic_sequence.TaskQueueName, worker.Options{})

	w.RegisterWorkflow(periodic_sequence.PeriodicSequencePolling)
	w.RegisterWorkflow(periodic_sequence.PollingChildWorkflow)
	testService := polling.NewTestService(50)
	activities := &periodic_sequence.PollingActivities{
		TestService: &testService,
	}
	w.RegisterActivity(activities)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
