package main

import (
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"

	child_workflow "github.com/temporalio/temporal-go-samples/child-workflow"
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
	defer c.Close()

	w := worker.New(c, "child-workflow", worker.Options{})

	w.RegisterWorkflow(child_workflow.SampleParentWorkflow)
	w.RegisterWorkflow(child_workflow.SampleChildWorkflow)

	err = w.Run()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}
}
