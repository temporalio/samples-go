package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"
)

const (
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
	flag.StringVar(&mode, "m", "", "Mode is worker or trigger.")
	flag.StringVar(&cron, "cron", "* * * * *", "Crontab schedule.")
	flag.Parse()

	switch mode {
	case "worker":
		w := startWorker()
		// The workers are supposed to be long running process that should not exit.
		// Use channel to wait for Ctrl+C.
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		// Stop worker, close connection, clean up resources.
		w.Stop()
	case "trigger":
		startWorkflow(cron)
	default:
		flag.PrintDefaults()
	}
}

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorker() worker.Worker {
	workerOptions := worker.Options{
		HostPort: client.DefaultHostPort,
		Logger:   logger,
	}

	w, err := worker.New(TaskListName, workerOptions)
	if err != nil {
		logger.Fatal("Unable to create worker", zap.Error(err))
	}

	w.RegisterWorkflow(SampleCronWorkflow)
	w.RegisterActivity(sampleCronActivity)

	err = w.Start()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}

	return w
}

// Start instance of the workflow.
func startWorkflow(cron string) {
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
		panic(err)
	}

	// This workflow ID can be user business logic identifier as well.
	workflowID := "cron_" + uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        TaskListName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
		CronSchedule:                    cron,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, SampleCronWorkflow)
	if err != nil {
		logger.Error("Unable to execute workflow", zap.Error(err))
	} else {
		logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
	}

	// Close connection, clean up resources.
	_ = c.CloseConnection()
}
