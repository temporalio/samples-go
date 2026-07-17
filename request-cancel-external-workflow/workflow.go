package requestcancelexternalworkflow

import (
	"errors"
	"go.temporal.io/sdk/workflow"
	"time"
)

// @@@SNIPSTART samples-go-cancellation-workflow-definition
// YourWorkflow is a Workflow Definition that shows how it can be canceled.
func CancellingWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    5 * time.Second,
		WaitForCancellation: true,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("cancel workflow started")
	var a *Activities // Used to call Activities by function pointer
	defer func() {

		if !errors.Is(ctx.Err(), workflow.ErrCanceled) {
			return
		}

		// When the Workflow is canceled, it has to get a new disconnected context to execute any Activities
		newCtx, _ := workflow.NewDisconnectedContext(ctx)
		err := workflow.ExecuteActivity(newCtx, a.CleanupActivity).Get(ctx, nil)
		if err != nil {
			logger.Error("CleanupActivity failed", "Error", err)
		}
	}()

	var cancelled string

	workflow.Go(ctx, func(ctx workflow.Context) {
		sigChan := workflow.GetSignalChannel(ctx, "cancelme")
		sel := workflow.NewSelector(ctx)
		sel.AddReceive(sigChan, func(c workflow.ReceiveChannel, more bool) {
			sigChan.Receive(ctx, &cancelled)
			logger.Info("received signal", cancelled)
		})

		sel.Select(ctx)
	})

	childCtx, _ := workflow.WithCancel(ctx)
	childCtx = workflow.WithChildOptions(childCtx, workflow.ChildWorkflowOptions{
		WaitForCancellation: true,
		WorkflowID:          "mykid",
	})
	childFuture := workflow.ExecuteChildWorkflow(childCtx, ChildWorkflow)

	if err := workflow.Await(ctx, func() bool {
		return cancelled == "cancelled"
	}); err != nil {
		return err
	}
	logger.Info("cancelling child")
	var childResult string

	if err := workflow.RequestCancelExternalWorkflow(ctx, "mykid", "").Get(ctx, &childResult); err != nil {
		return err
	}

	//cancelChild()
	if err := childFuture.Get(childCtx, &childResult); err != nil {
		logger.Error("child raised error", "err", err)
	}

	logger.Info("child result is", "result", childResult)

	logger.Info("Workflow Execution complete.")

	return nil
}
func ChildWorkflow(ctx workflow.Context) (myresult string, err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
		HeartbeatTimeout:    5 * time.Second,
		WaitForCancellation: true,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("CHILD: workflow started")
	myresult = "result: pristine"
	defer func() {
		logger.Info("CHILD: defer")
		if !errors.Is(ctx.Err(), workflow.ErrCanceled) {
			return
		}
		myresult = "result: child cancellation result"
		err = nil
		logger.Info("CHILD: cancel workflow")
	}()
	var neverSet bool
	if aerr := workflow.Await(ctx, func() bool {
		return neverSet
	}); aerr != nil {
		return "result: await errd", aerr
	}
	return "result: final value", nil
}
