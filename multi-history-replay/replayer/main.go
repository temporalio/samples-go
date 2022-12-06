package main

import (
	"context"
	"log"

	"github.com/temporalio/samples-go/helloworld"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

// GetWorkflows calls ListWorkflow with query and gets all workflow exection infos in a list.
func GetWorkflows(ctx context.Context, c client.Client, query string) ([]*workflowpb.WorkflowExecutionInfo, error) {
	var nextPageToken []byte
	var workflowExecutions []*workflowpb.WorkflowExecutionInfo
	for {
		resp, err := c.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			Query:         query,
			NextPageToken: nextPageToken,
		})
		if err != nil {
			return nil, err
		}
		workflowExecutions = append(workflowExecutions, resp.Executions...)
		nextPageToken = resp.NextPageToken
		if len(nextPageToken) == 0 {
			return workflowExecutions, nil
		}
	}
}

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
	ctx := context.Background()

	query := "WorkflowId='multiple_history_replay_workflowID'"
	log.Println("Listing workflows", "Query", query)
	workflowExecutions, err := GetWorkflows(ctx, c, query)
	if err != nil {
		log.Fatalln("Error listing workflows", err)
	}
	log.Println("Found workflows", "Count", len(workflowExecutions))

	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(helloworld.Workflow)
	for _, we := range workflowExecutions {
		execution := workflow.Execution{ID: we.Execution.GetWorkflowId(), RunID: we.Execution.GetRunId()}
		err = replayer.ReplayWorkflowExecution(ctx, c.WorkflowService(), nil, "default", execution)
		if err != nil {
			log.Fatalln("Error replaying history", err)
		}
	}
}
