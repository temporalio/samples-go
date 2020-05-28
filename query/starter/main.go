package main

import (
	"context"

	"go.temporal.io/temporal/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/query"
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
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:       "query_workflow",
		TaskList: "query",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, query.QueryWorkflow)
	if err != nil {
		logger.Fatal("Unable to execute workflow", zap.Error(err))
	}
	logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}
