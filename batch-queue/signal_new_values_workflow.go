package batch_queue

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"
	"time"

	"go.temporal.io/sdk/workflow"
)

func GetSignalNewValuesWorkflowID() string {
	return "SignalNewValuesWorkflowID"
}

// SignalNewValuesWorkflow is a workflow that keeps signaling new values to
// a forever-running workflow periodically. After each value is signaled, this
// workflow also writes it to a file. This file and the file where the batching
// workflow writes all values can be compared to see the durability.
func SignalNewValuesWorkflow(ctx workflow.Context) error {

	logger := workflow.GetLogger(ctx)
	logger.Info("SignalNewValuesWorkflow workflow started")

	var sendValue int
	sendValueFuture := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return rand.IntN(1_000)
	})
	err := sendValueFuture.Get(&sendValue)
	if err != nil {
		return err
	}

	// Note that to durably send, you can use activities. Here we don't handle
	// that, just fail early.
	for range 1000 {

		var sendValue int
		sendValueFuture := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
			return rand.IntN(1_000)
		})
		err := sendValueFuture.Get(&sendValue)
		if err != nil {
			return err
		}

		// The workflow we're signaling does continue-as-new, so do not pass a
		// run ID here.
		//
		// We can also signal in an activity if we want retries.
		err = workflow.SignalExternalWorkflow(ctx, GetAccumulateAndBatchWorkflowID(), "", SIGNAL_READ_VALS, strconv.Itoa(sendValue)).Get(ctx, nil)
		if err != nil {
			return err
		}

		ao := workflow.ActivityOptions{
			StartToCloseTimeout: 1 * time.Minute,
		}
		actx := workflow.WithActivityOptions(ctx, ao)
		err = workflow.ExecuteActivity(actx, WriteValToFile, sendValue).Get(ctx, nil)
		if err != nil {
			return err
		}

		err = workflow.Sleep(ctx, 1*time.Second)
		if err != nil {
			return err
		}
	}

	// avoid big history size
	return workflow.NewContinueAsNewError(ctx, SignalNewValuesWorkflow)
}

func WriteValToFile(ctx context.Context, val int) error {
	// Write the values to this file. We can compare values here to values sent
	// to the workflow to see the durability.
	f, err := os.OpenFile("values_sent.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	f.WriteString(strconv.Itoa(val))
	f.WriteString("\n")

	return nil
}
