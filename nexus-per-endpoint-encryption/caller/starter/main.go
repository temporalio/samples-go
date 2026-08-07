package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/caller"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/encryption"
	"github.com/temporalio/samples-go/nexus-per-endpoint-encryption/service"
	"github.com/temporalio/samples-go/nexus/options"
)

func main() {
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	clientOptions.DataConverter = encryption.NewDataConverter()
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	run, err := c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:        "nexus-per-endpoint-encryption-" + time.Now().Format("20060102150405"),
		TaskQueue: caller.TaskQueue,
	}, caller.PerEndpointEncryptionWorkflow)
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}

	var result service.WorkflowOutput
	if err := run.Get(context.Background(), &result); err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}
	log.Printf("Workflow result: %+v", result)
}
