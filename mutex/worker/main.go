package main

import (
	"context"
	"os"
	"os/signal"

	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/mutex"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
		Logger:   logger,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer c.CloseConnection()

	w := worker.New(c, "mutex", worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), mutex.ClientContextKey, c),
	})

	w.RegisterActivity(mutex.SignalWithStartMutexWorkflowActivity)
	w.RegisterWorkflow(mutex.MutexWorkflow)
	w.RegisterWorkflow(mutex.SampleWorkflowWithMutex)

	err = w.Start()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}
	defer w.Stop()

	// The workers are supposed to be long running process that should not exit.
	waitCtrlC()
}

func waitCtrlC() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
