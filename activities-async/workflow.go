package activities_async

import (
	"go.temporal.io/sdk/workflow"
	"time"
)

func AsyncActivitiesWorkflow(ctx workflow.Context) (res string, err error) {
	selector := workflow.NewSelector(ctx)
	var res1 string
	var res2 string
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	fut1 := workflow.ExecuteActivity(ctx, SayHello, "Temporal")
	fut2 := workflow.ExecuteActivity(ctx, SayGoodbye, "Temporal")
	selector.AddFuture(fut1, func(future workflow.Future) {
		err = future.Get(ctx, &res1)
	})
	selector.AddFuture(fut2, func(future workflow.Future) {
		err = future.Get(ctx, &res2)
	})
	selector.Select(ctx)
	if err != nil {
		return
	}
	selector.Select(ctx)
	if err == nil {
		res = res1 + " It was great to meet you, but time has come. " + res2
	}
	return
}
