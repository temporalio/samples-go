package goroutine

import (
	"fmt"
	"math/rand"
	"time"

	"go.temporal.io/sdk/workflow"
)

/**
* This sample workflow demonstrates how to use multiple Temporal gorotinues (instead of native goroutine) to process a
* a sequence of activities in parallel.
* In Temporal workflow, you should create goroutines using workflow.Go method.
 */

// SampleGoroutineWorkflow workflow definition
func SampleGoroutineWorkflow(ctx workflow.Context, parallelism int) (results []string, err error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	for i := 0; i < parallelism; i++ {
		input1 := fmt.Sprint(i) // Should be outside lambda to be captured correctly
		// Start a goroutine in a workflow safe way
		workflow.Go(ctx, func(gCtx workflow.Context) {
			// It is important to use the context passed to the goroutine function
			// An attempt to use the enclosing context would lead to failure.
			var result1 string
			err = workflow.ExecuteActivity(gCtx, Step1, input1).Get(gCtx, &result1)
			if err != nil {
				// Very naive error handling. Only the last error will be returned by the workflow
				return
			}
			var result2 string
			err = workflow.ExecuteActivity(gCtx, Step2, result1).Get(gCtx, &result2)
			if err != nil {
				return
			}
			results = append(results, result2)
		})
	}

	// Wait for Goroutines to complete. Await blocks until the condition function returns true.
	// The function is evaluated on every workflow state change. Consider using `workflow.AwaitWithTimeout` to
	// limit duration of the wait.
	_ = workflow.Await(ctx, func() bool {
		return err != nil || len(results) == parallelism
	})
	return
}

func Step1(input string) (output string, err error) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	return input + ", Step1", nil
}

func Step2(input string) (output string, err error) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	return input + ", Step2", nil
}
