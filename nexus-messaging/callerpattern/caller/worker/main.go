package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus-messaging/callerpattern/caller"
)

func main() {
	// Connect to the caller's namespace. For a non-local setup, provide additional
	// client options such as HostPort and TLS credentials.
	c, err := client.Dial(client.Options{Namespace: "my-caller-namespace"})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, caller.CallerTaskQueue, worker.Options{})
	w.RegisterWorkflow(caller.CallerWorkflow)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
