package main

import (
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
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
		Logger:   logger,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer c.Close()

	workerOptions := worker.Options{
		EnableSessionWorker: true, // Important for a worker to participate in the session
	}
	w := worker.New(c, "fileprocessing", workerOptions)

	w.RegisterWorkflow(fileprocessing.SampleFileProcessingWorkflow)
	w.RegisterActivity(&fileprocessing.Activities{BlobStore: &fileprocessing.BlobStore{}})

	err = w.Run()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}
}
