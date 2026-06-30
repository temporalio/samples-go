package main

import (
	"flag"
	"log"
	"os"

	"github.com/nexus-rpc/sdk-go/nexus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus-operations/update-workflow/api"
	"github.com/temporalio/samples-go/nexus-operations/update-workflow/handler"
	"github.com/temporalio/samples-go/nexus-operations/update-workflow/options"
)

func main() {
	set := flag.NewFlagSet("nexus-update-op-handler-worker", flag.ExitOnError)
	fp := options.NewClientFlagParser(set)
	set.Parse(os.Args[1:])
	clientOptions, err := fp.ClientOptions()
	if err != nil {
		log.Fatalf("Invalid options: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, api.HandlerTaskQueueName, worker.Options{})

	svc := nexus.NewService(api.CounterUpdateServiceName)
	if err := svc.Register(handler.IncrOperation); err != nil {
		log.Fatalln("Unable to register operations", err)
	}
	w.RegisterNexusService(svc)
	w.RegisterWorkflow(handler.CounterWorkflow)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
