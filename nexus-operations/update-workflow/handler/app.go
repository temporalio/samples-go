package handler

import (
	"context"
	"math/rand"
	"time"

	"github.com/temporalio/samples-go/nexus-operations/update-workflow/api"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporalnexus"
	"go.temporal.io/sdk/workflow"
)

// IncrOperation exposes the Workflow Update as a Nexus operation(via StartUpdateWorkflow)
var IncrOperation = temporalnexus.MustNewTemporalOperation(
	temporalnexus.TemporalOperationOptions[api.Input, api.Output]{
		Name: api.IncrOperationName,
		Start: func(
			ctx context.Context,
			nc temporalnexus.NexusClient,
			input api.Input,
			options temporalnexus.StartTemporalOperationOptions,
		) (temporalnexus.TemporalOperationResult[api.Output], error) {
			return temporalnexus.StartUpdateWorkflow[api.Output](ctx, nc, client.UpdateWorkflowOptions{
				WorkflowID:   input.WorkflowID,
				UpdateName:   api.IncrUpdateName,
				Args:         []any{input.Incr},
				WaitForStage: client.WorkflowUpdateStageAccepted,
			})
		},
	},
)

// CounterWorkflow is a long-running workflow that handles incr requests and stays open until it
// receives the Done signal.
func CounterWorkflow(ctx workflow.Context) (int, error) {
	logger := workflow.GetLogger(ctx)
	count := 0

	if err := workflow.SetUpdateHandler(
		ctx,
		api.IncrUpdateName,
		func(ctx workflow.Context, incr int) (api.Output, error) {
			if incr <= 0 {
				incr = 1
			}
			count += incr
			newCount := count
			logger.Info("counter updated", "incr", incr, "newValue", newCount)
			var randomSeconds int
			if err := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
				return rand.Intn(6) // sleep upto 5 seconds
			}).Get(&randomSeconds); err != nil {
				logger.Error("unexpected error", err)
				return api.Output{}, err
			}
			if err := workflow.Sleep(ctx, time.Duration(randomSeconds)*time.Second); err != nil {
				logger.Error("unexpected error", err)
				return api.Output{}, err
			}
			return api.Output{NewCount: newCount}, nil
		},
	); err != nil {
		return 0, err
	}

	workflow.GetSignalChannel(ctx, api.DoneSignalName).Receive(ctx, nil)
	return count, nil
}
