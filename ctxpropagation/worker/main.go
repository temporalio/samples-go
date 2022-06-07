package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/opentracing"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/ctxpropagation"
)

func main() {
	// Set tracer which will be returned by opentracing.GlobalTracer().
	closer := ctxpropagation.SetJaegerGlobalTracer()
	defer func() { _ = closer.Close() }()

	// Create interceptor
	tracingInterceptor, err := opentracing.NewInterceptor(opentracing.TracerOptions{})
	if err != nil {
		log.Fatalf("Failed creating interceptor: %v", err)
	}

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort:           client.DefaultHostPort,
		ContextPropagators: []workflow.ContextPropagator{ctxpropagation.NewContextPropagator()},
		Interceptors:       []interceptor.ClientInterceptor{tracingInterceptor},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "ctx-propagation", worker.Options{
		EnableLoggingInReplay: true,
	})

	w.RegisterWorkflow(ctxpropagation.CtxPropWorkflow)
	w.RegisterActivity(ctxpropagation.SampleActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
