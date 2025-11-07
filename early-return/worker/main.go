package main

import (
	"log"

	"github.com/temporalio/samples-go/early-return"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, earlyreturn.TaskQueueName, worker.Options{})

	w.RegisterWorkflow(earlyreturn.Workflow)
	w.RegisterActivity(earlyreturn.CompleteTransaction)
	w.RegisterActivity(earlyreturn.CancelTransaction)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
