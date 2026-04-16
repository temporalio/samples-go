package main

import (
	"log"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus-messaging/ondemandpattern/handler"
	"github.com/temporalio/samples-go/nexus-messaging/ondemandpattern/service"
)

func main() {
	// Connect to the handler's target namespace. For a non-local setup, provide additional
	// client options such as HostPort and TLS credentials.
	c, err := client.Dial(client.Options{Namespace: "my-target-namespace"})
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
