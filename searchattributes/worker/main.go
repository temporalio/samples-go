package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/searchattributes"
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

	ctx := context.WithValue(context.Background(), searchattributes.ClientCtxKey, c)

	w := worker.New(c, "search-attributes", worker.Options{
		BackgroundActivityContext: ctx,
	})

	w.RegisterWorkflow(searchattributes.SearchAttributesWorkflow)
	w.RegisterActivity(searchattributes.ListExecutions)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
