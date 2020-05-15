package main

import (
	"context"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.uber.org/zap"

	child_workflow "github.com/temporalio/temporal-go-samples/child-workflow"
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
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer func() { _ = c.CloseConnection() }()

	// This workflow ID can be user business logic identifier as well.
	workflowID := "parent-workflow_" + uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:       workflowID,
		TaskList: "child-workflow",
	}

	workflowRun, err := c.ExecuteWorkflow(context.Background(), workflowOptions, child_workflow.SampleParentWorkflow)
	if err != nil {
		logger.Fatal("Unable to execute workflow", zap.Error(err))
	}
	logger.Info("Started workflow",
		zap.String("WorkflowID", workflowRun.GetID()), zap.String("RunID", workflowRun.GetRunID()))

	// Synchronously wait for the workflow completion. Behind the scenes the SDK performs a long poll operation.
	// If you need to wait for the workflow completion from another process use
	// Client.GetWorkflow API to get an instance of a WorkflowRun.
	var result string
	err = workflowRun.Get(context.Background(), &result)
	if err != nil {
		logger.Fatal("Failure getting workflow result", zap.Error(err))
	}
	logger.Info("Workflow result: %v", zap.String("result", result))
}
