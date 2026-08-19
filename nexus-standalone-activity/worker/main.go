// @@@SNIPSTART samples-go-nexus-standalone-activity-worker
package main

import (
	"log"

	"github.com/nexus-rpc/sdk-go/nexus"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus-standalone-activity/handler"
	"github.com/temporalio/samples-go/nexus-standalone-activity/service"
)

const taskQueue = "nexus-handler-queue"

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, taskQueue, worker.Options{})

	svc := nexus.NewService(service.GreetingServiceName)
	err = svc.Register(handler.GreetOperation)
	if err != nil {
		log.Fatalln("Unable to register operations", err)
	}
	w.RegisterNexusService(svc)
	w.RegisterActivity(handler.CreateGreetingActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}

// @@@SNIPEND
