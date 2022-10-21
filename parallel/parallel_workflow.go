package parallel

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

/**
 * This sample workflow executes multiple branches in parallel using workflow.Go() method.
 */

// SampleParallelWorkflow workflow definition

type resp struct {
	result string
	err    error
}

func SampleParallelWorkflow(ctx workflow.Context) ([]string, error) {
	logger := workflow.GetLogger(ctx)
	defer logger.Info("Workflow completed.")

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second,
	}

	ch := workflow.NewChannel(ctx)

	// create 2 temporal coroutines
	for i := 0; i < 2; i++ {
		j := i
		workflow.Go(ctx, func(ctx workflow.Context) {
			branch := fmt.Sprintf("branch%d", j+1)

			logger.Info("Goroutine started", "branch", branch)
			waitForInput(ctx, branch)

			var result string
			err := workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, ao), SampleActivity, branch).Get(ctx, &result)
			ch.Send(ctx, &resp{
				result: result,
				err:    err,
			})
			logger.Info("goroutine completed", "branch", branch)
		})
	}

	var results []string
	for i := 0; i < 2; i++ {
		fmt.Println("Waiting for response", i)
		var v interface{}
		ch.Receive(ctx, &v)
		response, ok := v.(*resp)
		if !ok {
			fmt.Println("Invalid response")
			continue
		}
		if response.err != nil {
			fmt.Println("Got error from response", response.err)
			continue
		}

		results = append(results, response.result)

		if i == 0 {
			var result string
			fmt.Println("### going to execute activity after-first-result")
			err := workflow.ExecuteActivity(workflow.WithActivityOptions(ctx, ao), SampleActivity, "after-first-result").Get(ctx, &result)
			if err != nil {
				fmt.Println("Got error from after-branch1 response", err)
			} else {
				fmt.Println("### done to execute activity after-first-result", result)
				results = append(results, result)
			}
		}
	}
	fmt.Println("### returning result", results)
	return results, nil
}

func SampleActivity(input string) (string, error) {
	name := "sampleActivity"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + input, nil
}

func waitForInput(ctx workflow.Context, signalName string) {
	workflow.GetLogger(ctx).Debug("Waiting for signal", "signalName", signalName)
	signalChan := workflow.GetSignalChannel(ctx, signalName)
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(signalChan, func(channel workflow.ReceiveChannel, _ bool) {
		var v string
		channel.Receive(ctx, &v)
	})
	selector.Select(ctx)
	workflow.GetLogger(ctx).Debug("Resuming", "signalName", signalName)
}
