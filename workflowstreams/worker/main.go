package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"

	streams "github.com/temporalio/samples-go/workflowstreams"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, streams.TaskQueue, worker.Options{})

	w.RegisterWorkflow(streams.OrderWorkflow)
	w.RegisterWorkflow(streams.PipelineWorkflow)
	w.RegisterWorkflow(streams.HubWorkflow)
	w.RegisterWorkflow(streams.TickerWorkflow)
	w.RegisterActivity(streams.ChargeCard)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
