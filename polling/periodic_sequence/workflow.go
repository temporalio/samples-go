package periodic_sequence

import (
	"go.temporal.io/sdk/workflow"
	"time"
)

func PeriodicSequencePolling(ctx workflow.Context, pollingInterval time.Duration) (string, error) {
	cwo := workflow.ChildWorkflowOptions{}
	ctx = workflow.WithChildOptions(ctx, cwo)
	params := ChildWorkflowParams{
		SingleWorkflowAttempts: 10,
		PollingInterval:        pollingInterval,
	}
	res := workflow.ExecuteChildWorkflow(ctx, PollingChildWorkflow, params)
	var result string
	err := res.Get(ctx, &result)
	return result, err

}
