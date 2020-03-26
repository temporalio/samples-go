package main

import (
	"context"
	"go.uber.org/zap"
	"time"

	"github.com/temporalio/temporal-go-samples/parallel"
	"go.temporal.io/temporal/client"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		panic(err)
	}
	defer func() { _ = c.CloseConnection() }()
	workflowOptions := client.StartWorkflowOptions{
		TaskList:                        "parallel-task-list",
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Second * 10,
	}
	ctx := context.Background()
	we, err := c.ExecuteWorkflow(ctx, workflowOptions, parallel.SampleParallelWorkflow)
	if err != nil {
		panic(err)
	}
	logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))

	// Wait for workflow completion. This is rarely needed in real use cases
	// when workflows are potentially long running
	var result []string
	err = we.Get(ctx, &result)
	if err != nil {
		panic(err)
	}
	logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}
