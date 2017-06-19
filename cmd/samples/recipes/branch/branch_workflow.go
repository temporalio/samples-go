package main

import (
	"fmt"
	"time"

	"go.uber.org/cadence"
)

/**
 * This sample workflow executes multiple branches in parallel. The number of branches is controlled by passed in parameter.
 */

const (
	// ApplicationName is the task list for this sample
	ApplicationName = "branchGroup"

	totalBranches = 3
)

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	cadence.RegisterWorkflow(SampleBranchWorkflow)
	cadence.RegisterActivity(sampleActivity)
}

// SampleBranchWorkflow workflow decider
func SampleBranchWorkflow(ctx cadence.Context) error {
	var futures []cadence.Future
	// starts activities in parallel
	ao := cadence.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = cadence.WithActivityOptions(ctx, ao)

	for i := 1; i <= totalBranches; i++ {
		activityInput := fmt.Sprintf("branch %d of %d.", i, totalBranches)
		future := cadence.ExecuteActivity(ctx, sampleActivity, activityInput)
		futures = append(futures, future)
	}

	// wait until all futures are done
	for _, future := range futures {
		future.Get(ctx, nil)
	}

	cadence.GetLogger(ctx).Info("Workflow completed.")

	return nil
}

func sampleActivity(input string) (string, error) {
	name := "sampleActivity"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + name, nil
}
