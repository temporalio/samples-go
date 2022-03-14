package main

import (
	"context"
	"log"

	grpcproxy "github.com/temporalio/samples-go/grpc-proxy"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: "localhost:8081",
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "grpcproxy_workflowID",
		TaskQueue: "grpcproxy",
	}

	// The workflow input "My Compressed Friend" will be compressed by the DataConverter before being sent to Temporal
	we, err := c.ExecuteWorkflow(
		context.Background(),
		workflowOptions,
		grpcproxy.Workflow,
		"My Compressed Friend",
	)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
