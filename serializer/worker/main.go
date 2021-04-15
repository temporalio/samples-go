package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/serializer"
)

func main() {
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})

	if err != nil {
		log.Fatal("Unable to create client")
	}
	defer c.Close()

	ctx := context.Background()
	w := worker.New(c, serializer.Task_Queue_Name, worker.Options{
		BackgroundActivityContext: ctx,
	})

	w.RegisterWorkflow(serializer.ResourceWorkflow)
	w.RegisterActivity(serializer.ProcessEvent)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
