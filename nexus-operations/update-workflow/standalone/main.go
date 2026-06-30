package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/nexus-operations/update-workflow/api"
	"github.com/temporalio/samples-go/nexus-operations/update-workflow/options"
)

func main() {
	set := flag.NewFlagSet("nexus-update-op-starter", flag.ExitOnError)
	fp := options.NewClientFlagParser(set)
	incrementAmount := set.Int("incr", 1, "increment amount, defaults to 1 if <= 0")
	set.Parse(os.Args[1:])
	clientOptions, err := fp.ClientOptions()
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	nexusClient, err := c.NewNexusClient(client.NexusClientOptions{
		Endpoint: api.EndpointName,
		Service:  api.CounterUpdateServiceName,
	})
	if err != nil {
		log.Fatalln("Unable to create Nexus client", err)
	}

	handle, err := nexusClient.ExecuteOperation(
		context.Background(),
		api.IncrOperationName,
		api.Input{WorkflowID: api.CounterWorkflowID, Incr: *incrementAmount},
		client.StartNexusOperationOptions{
			ID:      "standalone-update-op-" + time.Now().Format("20060102150405"),
			Summary: "Incrementing count",
		})
	if err != nil {
		log.Fatalln("Unable to execute update operation", err)
	}
	log.Println("Started update workflow operation", "OperationID", handle.GetID())

	var res api.Output
	if err := handle.Get(context.Background(), &res); err != nil {
		log.Fatalln("Unable to get update workflow operation result", err)
	}
	log.Println("Incr result:", res.NewCount)
}
