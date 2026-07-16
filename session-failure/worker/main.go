package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	sessionfailure "github.com/temporalio/samples-go/session-failure"
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

	workerOptions := worker.Options{
		EnableSessionWorker: true, // Important for a worker to participate in the session
	}
	w := worker.New(c, "session-failure", workerOptions)

	w.RegisterWorkflow(sessionfailure.SampleSessionFailureRecoveryWorkflow)
	w.RegisterActivity(&sessionfailure.Activities{})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
