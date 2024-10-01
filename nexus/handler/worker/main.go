// @@@SNIPSTART workerhandler-nexus
package main

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/temporalio/samples-go/nexus/handler"
	"github.com/temporalio/samples-go/nexus/options"
	"github.com/temporalio/samples-go/nexus/service"
)

const (
	taskQueue = "my-handler-task-queue"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, taskQueue, worker.Options{})
	service := nexus.NewService(service.HelloServiceName)
	err = service.Register(handler.EchoOperation, handler.HelloOperation)
	if err != nil {
		log.Fatalln("Unable to register operations", err)
	}
	w.RegisterNexusService(service)
	w.RegisterWorkflow(handler.HelloHandlerWorkflow)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
// @@@SNIPEND