package main

import (
	"context"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/mutex"
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

	// This workflow ID can be user business logic identifier as well.
	resourceID := uuid.New()
	workflow1Options := client.StartWorkflowOptions{
		ID:       "SampleWorkflow1WithMutex_" + uuid.New(),
		TaskList: "mutex",
	}

	workflow2Options := client.StartWorkflowOptions{
		ID:       "SampleWorkflow2WithMutex_" + uuid.New(),
		TaskList: "mutex",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflow1Options, mutex.SampleWorkflowWithMutex, resourceID)
	if err != nil {
		logger.Error("Unable to execute workflow1", zap.Error(err))
	} else {
		logger.Info("Started workflow1", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
	}

	we, err = c.ExecuteWorkflow(context.Background(), workflow2Options, mutex.SampleWorkflowWithMutex, resourceID)
	if err != nil {
		logger.Error("Unable to execute workflow2", zap.Error(err))
	} else {
		logger.Info("Started workflow2", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
	}
}
