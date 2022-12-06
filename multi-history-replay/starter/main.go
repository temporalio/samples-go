package main

import (
	"context"
	"fmt"
	"log"

	"github.com/temporalio/samples-go/helloworld"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
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

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()
	ctx := context.Background()

	numberOfWorkflows := 10
	histories := make(chan *history.History, numberOfWorkflows)

	for w := 0; w < numberOfWorkflows; w++ {
		go func(workflowID int) {
			workflowOptions := client.StartWorkflowOptions{
				ID:        fmt.Sprintf("multiple_history_replay_workflowID_%d", workflowID),
				TaskQueue: "multiple-history-replay",
			}

			we, err := c.ExecuteWorkflow(ctx, workflowOptions, helloworld.Workflow, "Temporal")
			if err != nil {
				log.Fatalln("Unable to execute workflow", err)
			}

			log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

			// Wait for the workflow to finish
			var result string
			err = we.Get(ctx, &result)
			if err != nil {
				log.Fatalln("Unable get workflow result", err)
			}

			log.Println("Workflow finished", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

			// Download histroy
			history, err := GetWorkflowHistory(ctx, c, we.GetID(), we.GetRunID())
			if err != nil {
				log.Fatalln("Unable to get workflow history", err)
			}

			// Send the history to be replayed
			histories <- history
		}(w)
	}

	replayer := worker.NewWorkflowReplayer()
	replayer.RegisterWorkflow(helloworld.Workflow)

	for w := 0; w < numberOfWorkflows; w++ {
		h := <-histories
		err := replayer.ReplayWorkflowHistory(nil, h)
		if err != nil {
			log.Fatalln("Error replaying history", err)
		}
	}
}
