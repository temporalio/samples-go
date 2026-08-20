package main

import (
	"log"
	"os"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/encryption"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/handler"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/service"
	"github.com/temporalio/samples-go/nexus/options"
)

const taskQueue = "nexus-encryption-handler-task-queue"

func main() {
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	clientOptions.DataConverter = encryption.NewDataConverter()
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	nexusService := nexus.NewService(service.Name)
	if err := nexusService.Register(handler.SyncOperation, handler.AsyncOperation); err != nil {
		log.Fatalln("Unable to register Nexus operations", err)
	}

	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterNexusService(nexusService)
	w.RegisterWorkflow(handler.AsyncOperationWorkflow)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
