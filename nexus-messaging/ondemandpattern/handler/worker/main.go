package main

import (
	"log"
	"os"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus-messaging/ondemandpattern/handler"
	"github.com/temporalio/samples-go/nexus-messaging/ondemandpattern/service"
	"github.com/temporalio/samples-go/nexus/options"
)

const handlerNamespace = "my-target-namespace"

func main() {
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	clientOptions.Namespace = handlerNamespace

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, handler.HandlerTaskQueue, worker.Options{})

	svc := nexus.NewService(service.ServiceName)
	err = svc.Register(
		handler.RunFromRemoteOperation,
		handler.GetLanguagesOperation,
		handler.GetLanguageOperation,
		handler.SetLanguageOperation,
		handler.ApproveOperation,
	)
	if err != nil {
		log.Fatalln("Unable to register operations", err)
	}
	w.RegisterNexusService(svc)
	w.RegisterWorkflow(handler.GreetingWorkflow)
	w.RegisterActivity(handler.GreetingActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
