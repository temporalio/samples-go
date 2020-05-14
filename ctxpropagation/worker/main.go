package main

import (
	"os"
	"os/signal"

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
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer func() { _ = c.CloseConnection() }()

	w := worker.New(c, "ctx-propagation", worker.Options{
		Logger:                logger,
		EnableLoggingInReplay: true,
	})
	defer w.Stop()

	w.RegisterWorkflow(ctxpropagation.CtxPropWorkflow)
	w.RegisterActivity(ctxpropagation.SampleActivity)

	err = w.Start()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}

	// The workers are supposed to be long running process that should not exit.
	waitCtrlC()
}

func waitCtrlC() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
