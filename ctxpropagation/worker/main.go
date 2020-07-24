package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/temporal-go-samples/ctxpropagation"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
		ContextPropagators: []workflow.ContextPropagator{
			ctxpropagation.NewContextPropagator(),
		},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "ctx-propagation", worker.Options{
		EnableLoggingInReplay: true,
	})

	w.RegisterWorkflow(ctxpropagation.CtxPropWorkflow)
	w.RegisterActivity(ctxpropagation.SampleActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
