package main

import (
	"context"
	"log"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/temporal-go-samples/ctxpropagation"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort:           client.DefaultHostPort,
		ContextPropagators: []workflow.ContextPropagator{ctxpropagation.NewContextPropagator()},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowID := "ctx-propagation_" + uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "ctx-propagation",
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxpropagation.PropagateKey, &ctxpropagation.Values{Key: "test", Value: "tested"})

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, ctxpropagation.CtxPropWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
