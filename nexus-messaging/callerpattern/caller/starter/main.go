package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/nexus-messaging/callerpattern/caller"
)

func main() {
	// Connect to the caller's namespace. For a non-local setup, provide additional
	// client options such as HostPort and TLS credentials.
	c, err := client.Dial(client.Options{Namespace: "my-caller-namespace"})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()
	workflowOptions := client.StartWorkflowOptions{
		ID:        "nexus-messaging-caller-workflow-" + time.Now().Format("20060102150405"),
		TaskQueue: caller.CallerTaskQueue,
	}

	wr, err := c.ExecuteWorkflow(ctx, workflowOptions, caller.CallerWorkflow, "default-user")
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", wr.GetID(), "RunID", wr.GetRunID())

	var result []string
	if err := wr.Get(ctx, &result); err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}
	for i, entry := range result {
		fmt.Printf("[%d] %s\n", i+1, entry)
	}
}
