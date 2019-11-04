package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/pkg/errors"

	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

/**
 * This multi choice sample workflow Execute different parallel branches based on the result of an activity.
 */

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	workflow.Register(MultiChoiceWorkflow)
	activity.Register(getBasketOrderActivity)
}

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
	err := workflow.ExecuteActivity(ctx, getBasketOrderActivity).Get(ctx, &choices)
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
			f = workflow.ExecuteActivity(ctx, orderAppleActivity, item)
		case orderChoiceBanana:
			f = workflow.ExecuteActivity(ctx, orderBananaActivity, item)
		case orderChoiceCherry:
			f = workflow.ExecuteActivity(ctx, orderCherryActivity, item)
		case orderChoiceOrange:
			f = workflow.ExecuteActivity(ctx, orderOrangeActivity, item)
		default:
			logger.Error("Unexpected order.", zap.String("Order", item))
			return errors.New("Invalid Choice")
		}
		futures = append(futures, f)
	}

	// wait until all items in the basket order are processed
	for _, future := range futures {
		future.Get(ctx, nil)
	}

	logger.Info("Workflow completed.")
	return nil
}

func getBasketOrderActivity(ctx context.Context) ([]string, error) {
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
