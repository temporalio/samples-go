package main

import (
	"context"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/expense"
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

	expenseID := uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:        "expense_" + expenseID,
		TaskQueue: "expense",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, expense.SampleExpenseWorkflow, expenseID)
	if err != nil {
		logger.Fatal("Unable to execute workflow", zap.Error(err))
	}
	logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))

}
