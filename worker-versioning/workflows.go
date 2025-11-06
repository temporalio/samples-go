package worker_versioning

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// AutoUpgradingWorkflowV1 will automatically move to the latest worker version. We'll be making
// changes to it, which must be replay safe.
//
// Note that generally you won't want or need to include a version number in your workflow name if
// you're using the worker versioning feature. This sample does it to illustrate changes to the
// same code over time - but really what we're demonstrating here is the evolution of what would
// have been one workflow definition.
func AutoUpgradingWorkflowV1(ctx workflow.Context) error {
	workflow.GetLogger(ctx).Info("Changing workflow v1 started.", "StartTime", workflow.Now(ctx))

	// This workflow will listen for signals from our starter, and upon each signal either run
	// an activity, or conclude execution.
	signalChan := workflow.GetSignalChannel(ctx, "do-next-signal")
	for {
		var signal string
		signalChan.Receive(ctx, &signal)

		if signal == "do-activity" {
			workflow.GetLogger(ctx).Info("Changing workflow v1 running activity")
			ao := workflow.ActivityOptions{
				StartToCloseTimeout: 10 * time.Second,
			}
			ctx1 := workflow.WithActivityOptions(ctx, ao)
			err := workflow.ExecuteActivity(ctx1, SomeActivity, "v1").Get(ctx1, nil)
			if err != nil {
				return err
			}
		} else {
			workflow.GetLogger(ctx).Info("Concluding workflow v1")
			return nil
		}
	}
}

// AutoUpgradingWorkflowV1b represents us having made *compatible* changes to
// AutoUpgradingWorkflowV1.
//
// The compatible changes we've made are:
//   - Altering the log lines
//   - Using the workflow.GetVersion API to properly introduce branching behavior while maintaining
//     compatibility
func AutoUpgradingWorkflowV1b(ctx workflow.Context) error {
	workflow.GetLogger(ctx).Info("Changing workflow v1b started.", "StartTime", workflow.Now(ctx))

	// This workflow will listen for signals from our starter, and upon each signal either run
	// an activity, or conclude execution.
	signalChan := workflow.GetSignalChannel(ctx, "do-next-signal")
	for {
		var signal string
		signalChan.Receive(ctx, &signal)

		if signal == "do-activity" {
			workflow.GetLogger(ctx).Info("Changing workflow v1b running activity")
			v := workflow.GetVersion(ctx, "DifferentActivity", workflow.DefaultVersion, 1)
			if v == workflow.DefaultVersion {
				ao := workflow.ActivityOptions{
					StartToCloseTimeout: 10 * time.Second,
				}
				ctx1 := workflow.WithActivityOptions(ctx, ao)
				// Note it is a valid compatible change to alter the input to an activity.
				// However, because we're using the GetVersion API, this branch will never be
				// taken.
				err := workflow.ExecuteActivity(ctx1, SomeActivity, "v1b").Get(ctx1, nil)
				if err != nil {
					return err
				}
			} else {
				ao := workflow.ActivityOptions{
					StartToCloseTimeout: 10 * time.Second,
				}
				ctx1 := workflow.WithActivityOptions(ctx, ao)
				err := workflow.ExecuteActivity(ctx1, SomeIncompatibleActivity, &IncompatibleActivityInput{
					CalledBy: "v1b",
					MoreData: "hello!",
				}).Get(ctx1, nil)
				if err != nil {
					return err
				}
			}
		} else {
			workflow.GetLogger(ctx).Info("Concluding workflow v1b")
			break
		}
	}

	return nil
}

// PinnedWorkflowV1 demonstrates a workflow that likely has a short lifetime, and we want to always
// stay pinned to the same version it began on.
//
// Note that generally you won't want or need to include a version number in your workflow name if
// you're using the worker versioning feature. This sample does it to illustrate changes to the
// same code over time - but really what we're demonstrating here is the evolution of what would
// have been one workflow definition.
func PinnedWorkflowV1(ctx workflow.Context) error {
	workflow.GetLogger(ctx).Info("Pinned Workflow v1 started.", "StartTime", workflow.Now(ctx))

	signalChan := workflow.GetSignalChannel(ctx, "do-next-signal")
	for {
		var signal string
		signalChan.Receive(ctx, &signal)
		if signal == "conclude" {
			break
		}
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)
	err := workflow.ExecuteActivity(ctx1, SomeActivity, "Pinned-v1").Get(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}

// PinnedWorkflowV2 has changes that would make it incompatible with v1, and aren't protected by
// a patch.
func PinnedWorkflowV2(ctx workflow.Context) error {
	workflow.GetLogger(ctx).Info("Pinned Workflow v1 started.", "StartTime", workflow.Now(ctx))

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)
	// Here we call an activity where we didn't before, which is an incompatible change.
	err := workflow.ExecuteActivity(ctx1, SomeActivity, "Pinned-v2").Get(ctx, nil)
	if err != nil {
		return err
	}

	signalChan := workflow.GetSignalChannel(ctx, "do-next-signal")
	for {
		var signal string
		signalChan.Receive(ctx, &signal)
		if signal == "conclude" {
			break
		}
	}

	// We've also changed the activity type here, another incompatible change
	err = workflow.ExecuteActivity(ctx1, SomeIncompatibleActivity, &IncompatibleActivityInput{
		CalledBy: "Pinned-v2",
		MoreData: "hi",
	}).Get(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}

func SomeActivity(ctx context.Context, calledBy string) (string, error) {
	activity.GetLogger(ctx).Info("SomeActivity executing", "called by", calledBy)
	return calledBy, nil
}

// SomeIncompatibleActivity represents the need to change the interface to SomeActivity. Perhaps
// we didn't realize we would need to pass additional data, and we change the string parameter to
// a struct. (Hint: It's a great idea to always start with structs for this reason, as they can
// be extended without breaking compatibility as long as you use a wire format that maintains
// compatibility.)
func SomeIncompatibleActivity(ctx context.Context, input IncompatibleActivityInput) error {
	activity.GetLogger(ctx).Info("SomeIncompatibleActivity executing",
		"called by", input.CalledBy, "more data", input.MoreData)
	return nil
}

type IncompatibleActivityInput struct {
	CalledBy string
	MoreData string
}
