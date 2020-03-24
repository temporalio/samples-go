package choice

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * This multi choice sample workflow Execute different parallel branches based on the result of an activity.
 */

// MultiChoiceWorkflow Workflow Decider.
func MultiChoiceWorkflow(ctx workflow.Context) error {
	// Get basket order.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var choices []string
	err := workflow.ExecuteActivity(ctx, GetBasketOrderActivity).Get(ctx, &choices)
	if err != nil {
		return err
	}
	logger := workflow.GetLogger(ctx)

	var futures []workflow.Future
	for _, item := range choices {
		// choose next activity based on order result
		var f workflow.Future
		switch item {
		case orderChoiceApple:
			f = workflow.ExecuteActivity(ctx, OrderAppleActivity, item)
		case orderChoiceBanana:
			f = workflow.ExecuteActivity(ctx, OrderBananaActivity, item)
		case orderChoiceCherry:
			f = workflow.ExecuteActivity(ctx, OrderCherryActivity, item)
		case orderChoiceOrange:
			f = workflow.ExecuteActivity(ctx, OrderOrangeActivity, item)
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

func GetBasketOrderActivity(ctx context.Context) ([]string, error) {
	var basket []string
	for _, item := range _orderChoices {
		// some random decision
		if rand.Float32() <= 0.65 {
			basket = append(basket, item)
		}
	}

	if len(basket) == 0 {
		basket = append(basket, _orderChoices[rand.Intn(len(_orderChoices))])
	}

	activity.GetLogger(ctx).Info("Get basket order.", zap.Strings("Orders", basket))
	return basket, nil
}
