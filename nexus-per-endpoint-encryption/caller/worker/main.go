package main

import (
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/caller"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/encryption"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/service"
	"github.com/temporalio/samples-go/nexus/options"
)

func main() {
	opts, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	opts.DataConverter = encryption.NewCallerDataConverter()
	c, err := client.Dial(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	w := worker.New(c, service.CallerTaskQueue, worker.Options{})
	w.RegisterWorkflow(caller.EncryptionWorkflow)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}
