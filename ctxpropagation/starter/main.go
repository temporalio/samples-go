package main

import (
	"context"
	"log"

	"github.com/opentracing/opentracing-go"
	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/ctxpropagation"
)

func main() {
	// Set tracer which will be returned by opentracing.GlobalTracer().
	closer := ctxpropagation.SetJaegerGlobalTracer()
	defer func() { _ = closer.Close() }()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort:           client.DefaultHostPort,
		Tracer:             opentracing.GlobalTracer(),
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
