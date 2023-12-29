package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"

	open_telemetry "github.com/temporalio/samples-go/open-telemetry"
	"go.temporal.io/sdk/contrib/opentelemetry"
)

func main() {
	ctx := context.Background()
	tracerProvider, err := open_telemetry.CreateTraceProvider(ctx)
	if err != nil {
		log.Fatalln("Unable to create trace provider", err)
	}
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Fatalln("Unable to shutdown trace provider", err)
		}
	}()

	tracer, err := opentelemetry.NewTracingInterceptor(opentelemetry.TracerOptions{
		Tracer: tracerProvider.Tracer("go-sdk"),
	})
	if err != nil {
		log.Fatalln("Unable to create open telemetry interceptor", err)
	}

	// The client is a heavyweight object that should be created only once per process.
	c, err := client.Dial(client.Options{
		HostPort:     client.DefaultHostPort,
		Interceptors: []interceptor.ClientInterceptor{tracer},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "open-telemetry", worker.Options{})

	w.RegisterWorkflow(open_telemetry.Workflow)
	w.RegisterWorkflow(open_telemetry.ChildWorkflow)
	w.RegisterActivity(open_telemetry.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
