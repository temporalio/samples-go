package main

import (
	"context"
	"flag"
	"log"
	"os"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/nexus-operations/update-workflow/api"
	"github.com/temporalio/samples-go/nexus-operations/update-workflow/handler"
	"github.com/temporalio/samples-go/nexus-operations/update-workflow/options"
)

func main() {
	set := flag.NewFlagSet("nexus-update-op-handler-starter", flag.ExitOnError)
	fp := options.NewClientFlagParser(set)
	set.Parse(os.Args[1:])
	clientOptions, err := fp.ClientOptions()
	if err != nil {
		log.Fatalf("Invalid options: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()
	we, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        api.CounterWorkflowID,
		TaskQueue: api.HandlerTaskQueueName,
	}, handler.CounterWorkflow)
	if err != nil {
		log.Fatalln("Unable to start counter workflow", err)
	}
	log.Println("Started counter workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	var finalCount int
	if err := we.Get(ctx, &finalCount); err != nil {
		log.Fatalln("Workflow failed with error", err)
	}
	log.Println("Final count: ", finalCount)
}
