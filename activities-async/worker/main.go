package main

import (
	activities_async "github.com/temporalio/samples-go/activities-async"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
	w := worker.New(c, "async-activities-task-queue", worker.Options{})
	w.RegisterWorkflow(activities_async.AsyncActivitiesWorkflow)

	w.RegisterActivity(activities_async.SayHello)
	w.RegisterActivity(activities_async.SayGoodbye)
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
