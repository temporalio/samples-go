package main

import (
	"log"

	"github.com/temporalio/samples-go/reqresp"
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

	w := worker.New(c, "reqresp", worker.Options{})

	w.RegisterWorkflow(reqresp.UppercaseWorkflow)
	w.RegisterActivity(reqresp.UppercaseActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
