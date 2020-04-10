package choice

import (
	"errors"
	"time"

	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * This multi choice sample workflow executes multiple parallel branches based on the result of an activity.
 */

// MultiChoiceWorkflow Workflow definition.
func MultiChoiceWorkflow(ctx workflow.Context) error {
	// Get basket order.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	var orderActivities *OrderActivities // Used to call activities by function pointer

	var choices []string
	err := workflow.ExecuteActivity(ctx, orderActivities.GetBasketOrder).Get(ctx, &choices)
	if err != nil {
		return err
	}
	logger := workflow.GetLogger(ctx)

	var futures []workflow.Future
	for _, item := range choices {
		// choose next activity based on order result
		var f workflow.Future
		switch item {
		case OrderChoiceApple:
			f = workflow.ExecuteActivity(ctx, orderActivities.OrderApple, item)
		case OrderChoiceBanana:
			f = workflow.ExecuteActivity(ctx, orderActivities.OrderBanana, item)
		case OrderChoiceCherry:
			f = workflow.ExecuteActivity(ctx, orderActivities.OrderCherry, item)
		case OrderChoiceOrange:
			f = workflow.ExecuteActivity(ctx, orderActivities.OrderOrange, item)
		default:
			logger.Error("Unexpected order.", zap.String("Order", item))
			return errors.New("invalid choice")
		}
		futures = append(futures, f)
	}

	// wait until all items in the basket order are processed
	for _, future := range futures {
		_ = future.Get(ctx, nil)
	}

	logger.Info("Workflow completed.")
	return nil
}
