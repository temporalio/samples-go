package main

import (
	"context"
	"log"
	"net/http"

	"samples-go/yourapp"

	"go.temporal.io/sdk/client"
)

/*
Use the [`Dial()`](https://pkg.go.dev/go.temporal.io/sdk/client#Dial) API available in the [`go.temporal.io/sdk/client`](https://pkg.go.dev/go.temporal.io/sdk/client) package to create a new [`Client`](https://pkg.go.dev/go.temporal.io/sdk/client#Client).

If you don't provide [`HostPort`](https://pkg.go.dev/go.temporal.io/sdk/internal#ClientOptions), the Client defaults the address and port number to `127.0.0.1:7233`, which are the ports of the development Cluster.

Set a custom Namespace name in the Namespace field on an instance of the Client Options.

Use the [`ConnectionOptions`](https://pkg.go.dev/go.temporal.io/sdk/client#ConnectionOptions) API to connect a Client with mTLS.
*/

func main() {
	// Create a Temporal Client to communicate with the Temporal Cluster.
	// A Temporal Client is a heavyweight object that should be created just once per process.
	temporalClient, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create Temporal Client", err)
	}
	defer temporalClient.Close()
	// Start an HTTP server and listen on /start
	http.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		startWorkflowHandler(w, r, temporalClient)
	})
	err = http.ListenAndServe(":8091", nil)
	if err != nil {
		log.Fatalln("Unable to run http server", err)
	}
}

func startWorkflowHandler(w http.ResponseWriter, r *http.Request, temporalClient client.Client) {
	// Set the options for the Workflow Execution.
	// A Task Queue must be specified.
	// A custom Workflow Id is highly recommended.
	workflowOptions := client.StartWorkflowOptions{
		ID:        "your-workflow-id",
		TaskQueue: "your-custom-task-queue-name",
	}
	// Use an object as your Workflow Function parameter.
	// Objects enable your Function signature to remain compatible if fields change.
	workflowParams := yourapp.YourWorkflowParam{
		WorkflowParamX: "Hello",
		WorkflowParamY: 0,
	}
	// Make the call to the Temporal Cluster to start the Workflow Execution.
	workflowExecution, err := temporalClient.ExecuteWorkflow(
		context.Background(),
		workflowOptions,
		yourapp.YourWorkflowDefinition,
		workflowParams,
	)
	if err != nil {
		log.Fatalln("Unable to execute the Workflow", err)
	}
	log.Println("Started Workflow!")
	log.Println("WorkflowID:", workflowExecution.GetID())
	log.Println("RunID:", workflowExecution.GetRunID())
}

/* @dac
id: how-to-connect-to-a-development-cluster-in-go
title: How to connect to a Temporal dev Cluster in Go
label: Connect to a dev Cluster
description: Use the Dial() method on the Temporal Client and omit setting any client options. If there is a local dev Cluster running, the Client will connect to it.
lines:
@dac */
