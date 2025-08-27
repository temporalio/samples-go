package main

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"log"

	"github.com/temporalio/samples-go/external-env-conf"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(externalenvconf.LoadProfile())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "hello-world", worker.Options{})

	w.RegisterWorkflow(externalenvconf.Workflow)
	w.RegisterActivity(externalenvconf.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
