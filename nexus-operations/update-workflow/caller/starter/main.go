package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/nexus-operations/update-workflow/api"
	"github.com/temporalio/samples-go/nexus-operations/update-workflow/caller"
	"github.com/temporalio/samples-go/nexus/options"
)

func main() {
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()
	workflowOptions := client.StartWorkflowOptions{
		ID:        "counter-update-caller-" + time.Now().Format("20060102150405"),
		TaskQueue: caller.TaskQueue,
	}
	input := api.Input{WorkflowID: api.CounterWorkflowID}

	log.Println("Invoking incr operation")
	wr, err := c.ExecuteWorkflow(ctx, workflowOptions, caller.UpdateRemoteCounterWorkflow, input)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", wr.GetID(), "RunID", wr.GetRunID())

	var out api.Output
	if err := wr.Get(ctx, &out); err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}
	log.Println("Counter new value", out.NewCount)
}
