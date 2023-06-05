package main

import (
	"log"

	"github.com/temporalio/samples-go/polling"
	"github.com/temporalio/samples-go/polling/infrequent"

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

	w := worker.New(c, infrequent.TaskQueueName, worker.Options{})

	w.RegisterWorkflow(infrequent.InfrequentPolling)
	testService := polling.NewTestService(5)
	activities := &infrequent.PollingActivities{
		TestService: &testService,
	}
	w.RegisterActivity(activities)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
