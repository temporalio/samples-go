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

func main() {
	opts, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	opts.DataConverter = encryption.NewEncryptedDataConverter()
	c, err := client.Dial(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	w := worker.New(c, service.EncryptedTaskQueue, worker.Options{})
	svc := nexus.NewService(service.Name)
	if err := svc.Register(handler.SyncOperation, handler.AsyncOperation); err != nil {
		log.Fatal(err)
	}
	w.RegisterNexusService(svc)
	w.RegisterWorkflow(handler.EchoWorkflow)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}
