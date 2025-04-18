package main

import (
	"log"

	"github.com/temporalio/samples-go/nondeterminism"
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

	w := worker.New(c, "nde", worker.Options{})

	w.RegisterWorkflow(nondeterminism.WorkflowChangingActivityName)
	w.RegisterActivity(nondeterminism.ActivityOriginalName)
	w.RegisterActivity(nondeterminism.ActivityNewName)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
