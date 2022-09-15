package helloworld

import (
	"context"
	"reflect"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	// TODO(cretz): Remove when tagged
	_ "go.temporal.io/sdk/contrib/tools/workflowcheck/determinism"
)

type ErrorCode string

const (
	ErrorCodeFoo ErrorCode = "foo-code"
)

type NonRetryableError struct {
	Err  string
	Code *ErrorCode
}

func (e *NonRetryableError) Error() string {
	return e.Err
}

func NewNonRetryableError(err string) error {
	nonRetryErr := NonRetryableError{
		Err: err,
	}

	return &nonRetryErr
}

func NewNonRetryableErrorWithCode(err string, code ErrorCode) error {
	nonRetryErr := NonRetryableError{
		Err:  err,
		Code: &code,
	}

	return &nonRetryErr
}

var (
	nonRetryableErrorType = reflect.TypeOf(NewNonRetryableError("")).Elem().Name()
)

// Workflow is a Hello World workflow definition.
func Workflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			NonRetryableErrorTypes: []string{nonRetryableErrorType},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	var result string
	err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("HelloWorld workflow completed.", "result", result)

	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	// return "Hello " + name + "!", nil
	return "", temporal.NewApplicationError("some err", "NonRetryableError", NewNonRetryableErrorWithCode("foo msg", ErrorCodeFoo))
}
