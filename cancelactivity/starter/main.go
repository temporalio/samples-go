package main

import (
	"context"
	"flag"

	"go.temporal.io/temporal/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/cancelactivity"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	var workflowID string
	flag.StringVar(&workflowID, "w", "workflowID-to-cancel", "w is the workflowID of the workflow to be canceled.")
	flag.Parse()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer c.CloseConnection()

	workflowOptions := client.StartWorkflowOptions{
		ID:       workflowID,
		TaskList: "cancel-activity",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, cancelactivity.Workflow)
	if err != nil {
		logger.Fatal("Unable to execute workflow", zap.Error(err))
	}
	logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}
