package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/parallel"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		panic(err)
	}
	defer c.Close()
	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "parallel",
	}
	ctx := context.Background()
	we, err := c.ExecuteWorkflow(ctx, workflowOptions, parallel.SampleParallelWorkflow)
	if err != nil {
		panic(err)
	}
	log.Println("Started workflow", we.GetID(), we.GetRunID())

	// Wait for workflow completion. This is rarely needed in real use cases
	// when workflows are potentially long running
	var result []string
	err = we.Get(ctx, &result)
	if err != nil {
		panic(err)
	}
	log.Println("Completed workflow", "result", result)
}
