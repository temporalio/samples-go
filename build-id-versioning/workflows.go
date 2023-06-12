package build_id_versioning

import (
	"context"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"time"
)

// SampleChangingWorkflowV1 is a workflow we'll be making changes to, and represents the first
// version
func SampleChangingWorkflowV1(ctx workflow.Context) error {
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

// SampleChangingWorkflowV1b represents us having made *compatible* changes to
// SampleChangingWorkflowV1.
//
// The compatible changes we've made are:
//   - Altering the log lines
//   - Using the workflow.GetVersion API to properly introduce branching behavior while maintaining
//     compatibility
func SampleChangingWorkflowV1b(ctx workflow.Context) error {
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

// SampleChangingWorkflowV2 is fully incompatible with the other workflows, since it alters the
// sequence of commands without using workflow.GetVersion.
func SampleChangingWorkflowV2(ctx workflow.Context) error {
	workflow.GetLogger(ctx).Info("Changing workflow v2 started.", "StartTime", workflow.Now(ctx))
	err := workflow.Sleep(ctx, 10*time.Second)
	if err != nil {
		return err

	}
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)
	err = workflow.ExecuteActivity(ctx1, SomeActivity, "v2").Get(ctx, nil)
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
