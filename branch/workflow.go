// @@@SNIPSTART samples-go-branch-workflow-definition
package branch

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/workflow"
)

// SampleBranchWorkflow is a Temporal Workflow Definition
// This Workflow Definition shows how to call multiple Activities in parallel.
// The number of branches is controlled by a passed in parameter.
func SampleBranchWorkflow(ctx workflow.Context, totalBranches int) (result []string, err error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("SampleBranchWorkflow begin")

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
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
// @@@SNIPEND
