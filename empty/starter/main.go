// @@@SNIPSTART samples-go-branch-workflow-execution-starter
package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/empty"
)

func main() {
	// The client is a heavyweight object that should be created only once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "empty",
	}
	ctx := context.Background()
	we, err := c.ExecuteWorkflow(ctx, workflowOptions, empty.EmptyWorkflow).Get(ctx)
	if err != nil {
		log.Fatalln("Failure starting workflow", err)
	}
	log.Println("Started Workflow Execution", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
// @@@SNIPEND
