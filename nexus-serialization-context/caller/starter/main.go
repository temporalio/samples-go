package main

import (
	"context"
	"log"
	"os"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/nexus-serialization-context/caller"
	"github.com/temporalio/samples-go/nexus-serialization-context/encryption"
	"github.com/temporalio/samples-go/nexus-serialization-context/service"
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
	ctx := context.Background()
	run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID: "nexus-encryption-" + uuid.NewString(), TaskQueue: service.CallerTaskQueue,
	}, caller.EncryptionWorkflow, "Hello from Nexus")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Started workflow: ID=%s RunID=%s", run.GetID(), run.GetRunID())
	var result service.WorkflowOutput
	if err := run.Get(ctx, &result); err != nil {
		log.Fatal(err)
	}
	for _, reply := range []struct {
		endpoint string
		result   service.EndpointResults
	}{
		{service.CompressedEncrypted1Endpoint, result.CompressedEncrypted1},
		{service.CompressedEncrypted2Endpoint, result.CompressedEncrypted2},
		{service.Encrypted1Endpoint, result.Encrypted1},
		{service.Encrypted2Endpoint, result.Encrypted2},
	} {
		log.Printf("%s: sync=%q async=%q", reply.endpoint, reply.result.Sync, reply.result.Async)
	}
}
