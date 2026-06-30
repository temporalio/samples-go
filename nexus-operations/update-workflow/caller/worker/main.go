package main

import (
	"flag"
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus-operations/update-workflow/caller"
	"github.com/temporalio/samples-go/nexus-operations/update-workflow/options"
)

func main() {
	set := flag.NewFlagSet("nexus-update-op-caller-worker", flag.ExitOnError)
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

	w := worker.New(c, caller.TaskQueue, worker.Options{})
	w.RegisterWorkflow(caller.UpdateRemoteCounterWorkflow)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
