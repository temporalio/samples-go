// @@@SNIPSTART samples-go-nexus-caller-starter
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/nexus/caller"
	"github.com/temporalio/samples-go/nexus/options"
	"github.com/temporalio/samples-go/nexus/service"
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
	runWorkflow(c, fmt.Sprintf("nexus-caller-echo-%d", time.Now().UnixNano()), caller.EchoCallerWorkflow, "Nexus Echo 👋")
	runWorkflow(
		c,
		fmt.Sprintf("nexus-caller-hello-%s-%s-%d", "Nexus", service.EN, time.Now().UnixNano()),
		caller.HelloCallerWorkflow,
		"Nexus",
		service.EN,
	)
}

func runWorkflow(c client.Client, workflowID string, workflow interface{}, args ...interface{}) {
	ctx := context.Background()
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: caller.TaskQueue,
	}

	wr, err := c.ExecuteWorkflow(ctx, workflowOptions, workflow, args...)
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

// @@@SNIPEND
