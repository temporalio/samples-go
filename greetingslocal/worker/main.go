package main

import (
	"log"

	"github.com/temporalio/samples-go/greetings"
	"github.com/temporalio/samples-go/greetingslocal"
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

	w := worker.New(c, "greetings", worker.Options{})

	w.RegisterWorkflow(greetingslocal.GreetingSample)
	activities := &greetings.Activities{Name: "Temporal", Greeting: "Hello"}
	w.RegisterActivity(activities)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
