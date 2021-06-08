package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	cw "github.com/temporalio/samples-go/child-workflow-continue-as-new"
)
// @@@SNIPSTART samples-go-cw-cas-worker-starter
func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "child-workflow-continue-as-new", worker.Options{})

	w.RegisterWorkflow(cw.SampleParentWorkflow)
	w.RegisterWorkflow(cw.SampleChildWorkflow)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
// @@@SNIPEND
