package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/recovery"
	"github.com/temporalio/samples-go/recovery/cache"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.WithValue(context.Background(), recovery.TemporalClientKey, c)
	ctx = context.WithValue(ctx, recovery.WorkflowExecutionCacheKey, cache.NewLRU(10))

	w := worker.New(c, "recovery", worker.Options{
		BackgroundActivityContext: ctx,
	})

	w.RegisterWorkflowWithOptions(recovery.RecoverWorkflow, workflow.RegisterOptions{Name: "RecoverWorkflow"})
	w.RegisterWorkflowWithOptions(recovery.TripWorkflow, workflow.RegisterOptions{Name: "TripWorkflow"})
	w.RegisterActivity(recovery.ListOpenExecutions)
	w.RegisterActivity(recovery.RecoverExecutions)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
