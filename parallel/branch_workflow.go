package parallel

import (
	"fmt"
	"time"

	"go.temporal.io/temporal/workflow"
)

/**
 * This sample workflow executes multiple branches in parallel. The number of branches is controlled by passed in parameter.
 */

// SampleBranchWorkflow workflow definition
func SampleBranchWorkflow(ctx workflow.Context, totalBranches int) (result []string, err error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var futures []workflow.Future
	for i := 1; i <= totalBranches; i++ {
		activityInput := fmt.Sprintf("branch %d of %d.", i, totalBranches)
		future := workflow.ExecuteActivity(ctx, SampleActivity, activityInput)
		futures = append(futures, future)
	}

	// accumulate results
	for _, future := range futures {
		var singleResult string
		err = future.Get(ctx, &singleResult)
		if err != nil {
			return
		}
		result = append(result, singleResult)
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")
	return
}

func SampleActivity(input string) (string, error) {
	name := "sampleActivity"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + name, nil
}
