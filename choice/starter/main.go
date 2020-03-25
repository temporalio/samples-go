package main

import (
	"context"
	"flag"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/choice"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	var sampleCase string
	flag.StringVar(&sampleCase, "c", "single", "Sample case to run [single|multi].")
	flag.Parse()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}

	var workflowID string
	var workflowFn interface{}
	if sampleCase == "multi" {
		workflowID = "multi_choice_" + uuid.New()
		workflowFn = choice.MultiChoiceWorkflow
	} else if sampleCase == "single" {
		workflowID = "single_" + uuid.New()
		workflowFn = choice.ExclusiveChoiceWorkflow
	} else {
		flag.PrintDefaults()
		return
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        "choice-task-list",
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, workflowFn)
	if err != nil {
		logger.Error("Unable to execute workflow", zap.Error(err))
	} else {
		logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
	}

	// Close connection, clean up resources.
	_ = c.CloseConnection()
}
