package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/nexus-messaging/ondemandpattern/caller"
	"github.com/temporalio/samples-go/nexus/options"
)

const callerNamespace = "my-caller-namespace"

func main() {
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	clientOptions.Namespace = callerNamespace

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()
	workflowOptions := client.StartWorkflowOptions{
		ID:        "nexus-messaging-caller-remote-workflow-" + time.Now().Format("20060102150405"),
		TaskQueue: caller.CallerTaskQueue,
	}

	wr, err := c.ExecuteWorkflow(ctx, workflowOptions, caller.CallerRemoteWorkflow)
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
