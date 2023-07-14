package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"

	otelworkflow "github.com/temporalio/samples-go/opentelemetry"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	otelworkflow.Setup(ctx)
	defer otelworkflow.Shutdown(ctx)

	options := client.Options{
		HostPort:     "localhost:7233",
		Interceptors: []interceptor.ClientInterceptor{otelworkflow.GetInterceptor()},
	}

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(options)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "otel", worker.Options{})

	w.RegisterWorkflow(otelworkflow.Workflow)
	w.RegisterActivity(otelworkflow.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
