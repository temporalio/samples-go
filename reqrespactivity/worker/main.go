package main

import (
	"log"

	"github.com/temporalio/samples-go/reqrespactivity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "reqrespactivity", worker.Options{})

	w.RegisterWorkflow(reqrespactivity.UppercaseWorkflow)
	w.RegisterActivity(reqrespactivity.UppercaseActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
