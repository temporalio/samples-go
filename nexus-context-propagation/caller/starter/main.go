package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/temporalio/samples-go/ctxpropagation"
	"github.com/temporalio/samples-go/nexus/caller" // NOTE: reusing the generic nexus caller workflow
	"github.com/temporalio/samples-go/nexus/options"
	"github.com/temporalio/samples-go/nexus/service"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"
)

func main() {
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	// Set up context propagators from workflow and non-workflow contexts.
	clientOptions.ContextPropagators = []workflow.ContextPropagator{ctxpropagation.NewContextPropagator()}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
	ctx := context.Background()
	workflowOptions := client.StartWorkflowOptions{
		ID:        fmt.Sprintf("nexus-context-propagation-hello-caller-%s-%s-%d", "Nexus", service.ES, time.Now().UnixNano()),
		TaskQueue: caller.TaskQueue,
	}

	ctx = context.WithValue(ctx, ctxpropagation.PropagateKey, ctxpropagation.Values{
		Key:   "caller-id",
		Value: "samples-go",
	})
	wr, err := c.ExecuteWorkflow(ctx, workflowOptions, caller.HelloCallerWorkflow, "Nexus", service.ES)
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
