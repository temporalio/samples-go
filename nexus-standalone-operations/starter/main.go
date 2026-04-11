// @@@SNIPSTART samples-go-nexus-standalone-operations-starter
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"

	"github.com/temporalio/samples-go/nexus/service"
)

// This sample demonstrates standalone Nexus operations — executing Nexus operations
// directly from client code without wrapping them in a workflow.

const endpointName = "nexus-standalone-operations-endpoint"

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
		Service:  service.HelloServiceName,
	})
	if err != nil {
		log.Fatalln("Unable to create Nexus client", err)
	}

	// Execute the sync Echo operation.
	echoHandle, err := nexusClient.ExecuteOperation(context.Background(), service.EchoOperationName, service.EchoInput{Message: "hello"}, client.StartNexusOperationOptions{
		ID:                     "nexus-standalone-echo-op",
		ScheduleToCloseTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatalln("Unable to execute Echo operation", err)
	}
	log.Println("Started Echo operation", "OperationID", echoHandle.GetID())

	var echoResult service.EchoOutput
	err = echoHandle.Get(context.Background(), &echoResult)
	if err != nil {
		log.Fatalln("Unable to get Echo operation result", err)
	}
	log.Println("Echo result:", echoResult.Message)

	// Execute the async (workflow-backed) Hello operation.
	helloHandle, err := nexusClient.ExecuteOperation(context.Background(), service.HelloOperationName, service.HelloInput{Name: "Temporal", Language: service.EN}, client.StartNexusOperationOptions{
		ID:                     "nexus-standalone-hello-op",
		ScheduleToCloseTimeout: 10 * time.Second,
	})
	if err != nil {
		log.Fatalln("Unable to execute Hello operation", err)
	}
	log.Println("Started Hello operation", "OperationID", helloHandle.GetID())

	var helloResult service.HelloOutput
	err = helloHandle.Get(context.Background(), &helloResult)
	if err != nil {
		log.Fatalln("Unable to get Hello operation result", err)
	}
	log.Println("Hello result:", helloResult.Message)

	// List Nexus operations using the base client (not NexusClient).
	listResp, err := c.ListNexusOperations(context.Background(), client.ListNexusOperationsOptions{
		Query: fmt.Sprintf("Endpoint = '%s'", endpointName),
	})
	if err != nil {
		log.Fatalln("Unable to list Nexus operations", err)
	}

	log.Println("ListNexusOperations results:")
	for metadata, err := range listResp.Results {
		if err != nil {
			log.Fatalln("Error iterating operations", err)
		}
		log.Printf("\tOperationID: %s, Operation: %s, Status: %v\n",
			metadata.OperationID, metadata.Operation, metadata.Status)
	}

	// Count Nexus operations using the base client (not NexusClient).
	countResp, err := c.CountNexusOperations(context.Background(), client.CountNexusOperationsOptions{
		Query: fmt.Sprintf("Endpoint = '%s'", endpointName),
	})
	if err != nil {
		log.Fatalln("Unable to count Nexus operations", err)
	}
	log.Println("Total Nexus operations:", countResp.Count)
}

// @@@SNIPEND
