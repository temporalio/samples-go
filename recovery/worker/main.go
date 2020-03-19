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
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}

	ctx := context.WithValue(context.Background(), recovery.TemporalClientKey, c)
	ctx = context.WithValue(ctx, recovery.WorkflowExecutionCacheKey, cache.NewLRU(10))

	workflowWorker := worker.New(c, "recovery-task-list", worker.Options{
		Logger:                    logger,
		BackgroundActivityContext: ctx,
	})

	workflowWorker.RegisterWorkflowWithOptions(recovery.RecoverWorkflow, workflow.RegisterOptions{Name: "RecoverWorkflow"})
	workflowWorker.RegisterWorkflowWithOptions(recovery.TripWorkflow, workflow.RegisterOptions{Name: "TripWorkflow"})

	err = workflowWorker.Start()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}

	hostWorker := worker.New(c, recovery.HostID, worker.Options{
		Logger:                    logger,
		BackgroundActivityContext: ctx,
		DisableWorkflowWorker:     true,
	})

	hostWorker.RegisterActivity(recovery.ListOpenExecutions)
	hostWorker.RegisterActivity(recovery.RecoverExecutions)

	err = hostWorker.Start()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}

	// The workers are supposed to be long running process that should not exit.
	waitCtrlC()
	// Stop worker, close connection, clean up resources.
	hostWorker.Stop()
	workflowWorker.Stop()
	_ = c.CloseConnection()
}

func waitCtrlC() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
