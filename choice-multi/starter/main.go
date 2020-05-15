package main

import (
	"context"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.uber.org/zap"

	choice_multi "github.com/temporalio/temporal-go-samples/choice-multi"
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

	workflowOptions := client.StartWorkflowOptions{
		ID:       "multi_choice_" + uuid.New(),
		TaskList: "choice-multi",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, choice_multi.MultiChoiceWorkflow)
	if err != nil {
		logger.Fatal("Unable to execute workflow", zap.Error(err))
	}
	logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))

}
