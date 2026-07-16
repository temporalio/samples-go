package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/nexus-cancelation/caller"
	"github.com/temporalio/samples-go/nexus/options"
)

func main() {
	ctx := context.Background()
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "nexus_hello_caller_workflow_" + time.Now().Format("20060102150405"),
		TaskQueue: caller.TaskQueue,
	}

	wr, err := c.ExecuteWorkflow(ctx, workflowOptions, caller.HelloCallerWorkflow, "Nexus")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", wr.GetID(), "RunID", wr.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = wr.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
