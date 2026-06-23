package main

import (
	"context"
	"log"
	"os"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/nexus-activity-operation/options"
	"github.com/temporalio/samples-go/nexus-activity-operation/service"
)

const endpointName = "my-nexus-endpoint-name"

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

	// client.NewNexusClient binds a Nexus client to a single endpoint + service. Use this for
	// standalone Nexus calls from non-workflow code; from inside a workflow, use
	// workflow.NewNexusClient instead.
	nc, err := c.NewNexusClient(client.NexusClientOptions{
		Endpoint: endpointName,
		Service:  service.HelloServiceName,
	})
	if err != nil {
		log.Fatalln("Unable to create Nexus client", err)
	}

	ctx := context.Background()

	// Activity-backed Nexus operation. ExecuteOperation returns once the operation has been
	// started; handle.Get blocks until the backing activity completes.
	handle, err := nc.ExecuteOperation(ctx, service.HelloOperationName, service.HelloInput{Name: "Nexus", Language: service.ES}, client.StartNexusOperationOptions{
		ID: "hello-op",
	})
	if err != nil {
		log.Fatalln("Unable to start hello operation", err)
	}
	var out service.HelloOutput
	if err := handle.Get(ctx, &out); err != nil {
		log.Fatalln("Unable to get hello result", err)
	}
	log.Println("Hello result:", out.Message)
}
