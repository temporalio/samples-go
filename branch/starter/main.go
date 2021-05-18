// @@@SNIPSTART samples-go-branch-workflow-execution-starter
package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/branch"
)

func main() {
	// The client is a heavyweight object that should be created only once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "branch",
	}
	ctx := context.Background()
	we, err := c.ExecuteWorkflow(ctx, workflowOptions, branch.SampleBranchWorkflow, 10)
	if err != nil {
		log.Fatalln("Failure starting workflow", err)
	}
	log.Println("Started Workflow Execution", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Wait for Workflow Execution completion.
	// This is rarely needed in real use cases when Workflow Executions are potentially long running.
	var result []string
	err = we.Get(ctx, &result)
	if err != nil {
		panic(err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
// @@@SNIPEND
