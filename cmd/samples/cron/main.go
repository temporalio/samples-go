package main

import (
	"context"
	"flag"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"
)

const (
	// DomainName is the name of domain where workflow will be created.
	DomainName = "samples-domain"

	// TaskListName is the task list for this sample.
	TaskListName = "cron-task-list"
)

var (
	logger *zap.Logger
)

func main() {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	var mode string
	var cron string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.StringVar(&cron, "cron", "* * * * *", "Crontab schedule. Default \"* * * * *\"")
	flag.Parse()

	switch mode {
	case "worker":
		startWorker()
		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by Ctrl+C.
		select {}
	case "trigger":
		startWorkflow(cron)
	}
}

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorker() {
	workerOptions := worker.Options{
		Logger: logger,
	}

	worker, err := worker.New(DomainName, TaskListName, workerOptions)
	if err != nil {
		logger.Fatal("Unable to create worker", zap.Error(err))
	}

	worker.RegisterWorkflow(SampleCronWorkflow)
	worker.RegisterActivity(sampleCronActivity)

	err = worker.Start()
	if err != nil {
		logger.Error("Unable to start worker", zap.Error(err))
	}
}

// Start instance of the workflow.
func startWorkflow(cron string) {
	// This workflow ID can be user business logic identifier as well.
	workflowID := "cron_" + uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        TaskListName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
		CronSchedule:                    cron,
	}

	client, err := client.NewClient(DomainName, client.Options{})
	if err != nil {
		logger.Fatal("Unable to build client", zap.Error(err))
		panic(err)
	}

	we, err := client.ExecuteWorkflow(context.Background(), workflowOptions, SampleCronWorkflow)
	if err != nil {
		logger.Error("Unable to execute workflow", zap.Error(err))
	} else {
		logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
	}
}
