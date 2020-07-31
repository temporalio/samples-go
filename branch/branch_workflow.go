package branch

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

// SampleBranchWorkflow workflow definition
// This workflow executes multiple activities in parallel. The number of branches is controlled by a passed in parameter.
func SampleBranchWorkflow(ctx workflow.Context, totalBranches int) (result []string, err error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("SampleBranchWorkflow begin")

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
	logger.Info("Activities started")

	// accumulate results
	for _, future := range futures {
		var singleResult string
		err = future.Get(ctx, &singleResult)
		logger.Info("Activity returned with result", "resutl", singleResult)
		if err != nil {
			return
		}
		result = append(result, singleResult)
	}

	logger.Info("SampleBranchWorkflow end")
	return
}

func SampleActivity(input string) (string, error) {
	name := "sampleActivity"
	fmt.Printf("Run %s with input %v \n", name, input)
	return "Result_" + input, nil
}
