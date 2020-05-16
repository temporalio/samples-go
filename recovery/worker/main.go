package main

import (
	"context"
	"os"
	"os/signal"

	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/recovery"
	"github.com/temporalio/temporal-go-samples/recovery/cache"
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

	ctx := context.WithValue(context.Background(), recovery.TemporalClientKey, c)
	ctx = context.WithValue(ctx, recovery.WorkflowExecutionCacheKey, cache.NewLRU(10))

	w := worker.New(c, "recovery", worker.Options{
		BackgroundActivityContext: ctx,
	})

	w.RegisterWorkflowWithOptions(recovery.RecoverWorkflow, workflow.RegisterOptions{Name: "RecoverWorkflow"})
	w.RegisterWorkflowWithOptions(recovery.TripWorkflow, workflow.RegisterOptions{Name: "TripWorkflow"})
	w.RegisterActivity(recovery.ListOpenExecutions)
	w.RegisterActivity(recovery.RecoverExecutions)

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
