// @@@SNIPSTART samples-go-nexus-standalone-activity-starter
package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"

	"github.com/temporalio/samples-go/nexus-standalone-activity/service"
)

// Executes the Activity-backed Nexus Operation from client code. The operation is standalone: it is
// started directly by this client rather than from within a caller Workflow.

const endpointName = "my-nexus-endpoint"

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// Create a NexusClient bound to the endpoint and service.
	// The endpoint must be pre-created on the server (see README).
	nexusClient, err := c.NewNexusClient(client.NexusClientOptions{
		Endpoint: endpointName,
		Service:  service.GreetingServiceName,
	})
	if err != nil {
		log.Fatalln("Unable to create Nexus client", err)
	}

	handle, err := nexusClient.ExecuteOperation(context.Background(), service.GreetOperationName, service.GreetingInput{Name: "World"}, client.StartNexusOperationOptions{
		ID:                     "greeting-" + uuid.NewString(),
		ScheduleToCloseTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatalln("Unable to execute Greet operation", err)
	}
	log.Println("Started Greet operation", "OperationID", handle.GetID())

	var result service.GreetingOutput
	err = handle.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable to get Greet operation result", err)
	}
	log.Println(result.Message)
}

// @@@SNIPEND
