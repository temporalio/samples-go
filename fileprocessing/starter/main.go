package main

import (
	"context"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/fileprocessing"
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

	fileID := uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "fileprocessing_" + fileID,
		TaskList:                        "fileprocessing",
		ExecutionStartToCloseTimeout:    time.Minute,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, fileprocessing.SampleFileProcessingWorkflow, fileID)
	if err != nil {
		logger.Fatal("Unable to execute workflow", zap.Error(err))
	}
	logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))

}
