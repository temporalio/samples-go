package fileprocessing

import (
	"errors"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var (
	ErrSessionHostDown = errors.New("session host down")
)

// SampleSessionFailureRecoveryWorkflow workflow definition
func SampleSessionFailureRecoveryWorkflow(ctx workflow.Context, fileName string) (err error) {

	err = runSession(ctx, fileName)
	numOfRetries := 10
	for err != nil && numOfRetries >= 0 {
		// Only retry if we detected the session failed. In a production application
		// it may make sense to also retry if some other errors occur, it
		// depends on your business logic.
		if errors.Is(err, ErrSessionHostDown) {
			workflow.Sleep(ctx, 5*time.Minute)
			err = runSession(ctx, fileName)
		} else {
			break
		}
		numOfRetries--
	}

	if err != nil {
		workflow.GetLogger(ctx).Error("Workflow failed.", "Error", err.Error())
	} else {
		workflow.GetLogger(ctx).Info("Workflow completed.")
	}
	return err
}

func runSession(ctx workflow.Context, fileName string) (err error) {

	so := &workflow.SessionOptions{
		CreationTimeout:  time.Minute,
		ExecutionTimeout: 20 * time.Minute,
	}
	sessionCtx, err := workflow.CreateSession(ctx, so)
	if err != nil {
		// In a production application you may want to distinguish between not being able to create
		// a session and a host going down.
		if temporal.IsTimeoutError(err) {
			workflow.GetLogger(ctx).Error("Session failed", "Error", err.Error())
			err = ErrSessionHostDown
		}
		return err
	}

	defer func() {
		workflow.CompleteSession(sessionCtx)
		// If the session host fails any scheduled activity started on the host will be cancelled.
		//
		// Note: SessionState is inherently a stale view of the session state see the README.md of
		// this sample for more details
		if workflow.GetSessionInfo(sessionCtx).SessionState == workflow.SessionStateFailed {
			err = ErrSessionHostDown
		}
	}()

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		// When running an activity in a session you don't need to specify a heartbeat timeout to
		// detect the host going down, the session heartbeat timeout will handle that for you.
		// You may still want to specify a heartbeat timeout if the activity can get stuck or
		// you want to record progress with the heartbeat details.
		HeartbeatTimeout: 40 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
		},
	}
	sessionCtx = workflow.WithActivityOptions(sessionCtx, ao)

	var a *Activities
	err = workflow.ExecuteActivity(sessionCtx, a.PrepareWorkerActivity).Get(sessionCtx, nil)
	if err != nil {
		return err
	}

	return workflow.ExecuteActivity(sessionCtx, a.LongRunningActivity).Get(sessionCtx, nil)
}
