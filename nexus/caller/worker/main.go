package main

import (
	"log"
	"os"

	"github.com/temporalio/samples-go/nexus/caller"
	"github.com/temporalio/samples-go/nexus/options"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, caller.TaskQueue, worker.Options{})

	w.RegisterWorkflow(caller.EchoCallerWorkflow)
	w.RegisterWorkflow(caller.HelloCallerWorkflow)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
