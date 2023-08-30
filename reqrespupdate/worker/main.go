package main

import (
	"log"

	"github.com/temporalio/samples-go/reqrespupdate"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "reqrespupdate", worker.Options{})

	w.RegisterWorkflow(reqrespupdate.UppercaseWorkflow)
	w.RegisterActivity(reqrespupdate.UppercaseActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
