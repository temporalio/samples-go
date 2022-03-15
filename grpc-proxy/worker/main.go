package main

import (
	"log"

	grpcproxy "github.com/temporalio/samples-go/grpc-proxy"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: "localhost:8081",
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "grpcproxy", worker.Options{})

	w.RegisterWorkflow(grpcproxy.Workflow)
	w.RegisterActivity(grpcproxy.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
