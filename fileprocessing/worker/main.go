package main

import (
	"os"
	"os/signal"

	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/fileprocessing"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}

	workerOptions := worker.Options{
		Logger:                logger,
		EnableLoggingInReplay: true,
		EnableSessionWorker:   true, // !!! Must be enabled for sessions to work
	}
	w := worker.New(c, "fileprocessing-task-list", workerOptions)

	w.RegisterWorkflow(fileprocessing.SampleFileProcessingWorkflow)
	w.RegisterActivityWithOptions(fileprocessing.DownloadFileActivity, activity.RegisterOptions{Name: fileprocessing.DownloadFileActivityName})
	w.RegisterActivityWithOptions(fileprocessing.ProcessFileActivity, activity.RegisterOptions{Name: fileprocessing.ProcessFileActivityName})
	w.RegisterActivityWithOptions(fileprocessing.UploadFileActivity, activity.RegisterOptions{Name: fileprocessing.UploadFileActivityName})

	err = w.Start()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}

	// The workers are supposed to be long running process that should not exit.
	waitCtrlC()
	// Stop the worker, close connection, clean up resources.
	w.Stop()
	_ = c.CloseConnection()
}

func waitCtrlC() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
