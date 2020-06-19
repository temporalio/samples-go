package main

import (
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/ctxpropagation"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
		ContextPropagators: []workflow.ContextPropagator{
			ctxpropagation.NewContextPropagator(),
		},
		Logger: logger,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer c.Close()

	w := worker.New(c, "ctx-propagation", worker.Options{
		EnableLoggingInReplay: true,
	})

	w.RegisterWorkflow(ctxpropagation.CtxPropWorkflow)
	w.RegisterActivity(ctxpropagation.SampleActivity)

	err = w.Run()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}
}
