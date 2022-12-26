package main

import (
	capped_activities "github.com/temporalio/samples-go/capped-concurrent-activities"
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

	w := worker.New(c, "capped-activities", worker.Options{})

	w.RegisterWorkflow(capped_activities.CappedActivitiesWorkflow)
	w.RegisterActivity(capped_activities.ChunkProcessingActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
