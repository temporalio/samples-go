package main

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/nexus-rpc/sdk-go/nexus"
	"github.com/temporalio/samples-go/nexus-activity-operation/handler"
	"github.com/temporalio/samples-go/nexus-activity-operation/options"
	"github.com/temporalio/samples-go/nexus-activity-operation/service"
)

const (
	taskQueue = "my-handler-task-queue"
)

func main() {
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
	svc := nexus.NewService(service.HelloServiceName)
	if err := svc.Register(handler.HelloOperation); err != nil {
		log.Fatalln("Unable to register operations", err)
	}
	w.RegisterNexusService(svc)
	w.RegisterActivity(handler.HelloHandlerActivity)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
