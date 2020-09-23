package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/fileprocessing"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workerOptions := worker.Options{
		EnableSessionWorker: true, // Important for a worker to participate in the session
	}
	w := worker.New(c, "fileprocessing", workerOptions)

	w.RegisterWorkflow(fileprocessing.SampleFileProcessingWorkflow)
	w.RegisterActivity(&fileprocessing.Activities{BlobStore: &fileprocessing.BlobStore{}})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
