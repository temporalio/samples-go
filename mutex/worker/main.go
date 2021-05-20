package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/mutex"
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

	w := worker.New(c, "mutex", worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), mutex.ClientContextKey, c),
	})

	w.RegisterActivity(mutex.SignalWithStartMutexWorkflowActivity)
	w.RegisterWorkflow(mutex.MutexWorkflow)
	w.RegisterWorkflow(mutex.SampleWorkflowWithMutex)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
