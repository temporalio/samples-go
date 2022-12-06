package main

import (
	"context"
	"log"

	"github.com/temporalio/samples-go/helloworld"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func GetWorkflowHistory(ctx context.Context, client client.Client, id, runID string) (*history.History, error) {
	var hist history.History
	iter := client.GetWorkflowHistory(ctx, id, runID, false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return nil, err
		}
		hist.Events = append(hist.Events, event)
	}
	return &hist, nil
}

// GetWorkflows calls ListWorkflow with query and gets as many workflow executions until there are no more or we the number exceeds maxWorkflows
func GetWorkflows(ctx context.Context, c client.Client, query string, maxWorkflows int) ([]*workflow.WorkflowExecutionInfo, error) {
	var nextPageToken []byte
	workflowExecutions := make([]*workflow.WorkflowExecutionInfo, 0)
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
		if nextPageToken == nil || len(workflowExecutions) >= maxWorkflows {
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
	workflowExecutions, err := GetWorkflows(ctx, c, query, 10)
	if err != nil {
		log.Fatalln("Error listing workflows", err)
	}
	log.Println("Found workflows", "Count", len(workflowExecutions))

	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(helloworld.Workflow)
	for _, we := range workflowExecutions {
		histroy, err := GetWorkflowHistory(ctx, c, we.Execution.GetWorkflowId(), we.Execution.GetRunId())
		if err != nil {
			log.Fatalln("Error getting history", err)
		}
		err = replayer.ReplayWorkflowHistory(nil, histroy)
		if err != nil {
			log.Fatalln("Error replaying history", err)
		}
	}
}
