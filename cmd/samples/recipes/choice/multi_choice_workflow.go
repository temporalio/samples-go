package main

import (
	"context"
	"math/rand"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/cadence"
	"go.uber.org/zap"
)

/**
 * This multi choice sample workflow Execute different parallel branches based on the result of an activity.
 */

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	cadence.RegisterWorkflow(MultiChoiceWorkflow)
	cadence.RegisterActivity(getBasketOrderActivity)
}

// MultiChoiceWorkflow Workflow Decider.
func MultiChoiceWorkflow(ctx cadence.Context) error {
	// Get basket order.
	ao := cadence.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = cadence.WithActivityOptions(ctx, ao)

	var choices []string
	err := cadence.ExecuteActivity(ctx, getBasketOrderActivity).Get(ctx, &choices)
	if err != nil {
		return err
	}
	logger := cadence.GetLogger(ctx)

	var futures []cadence.Future
	for _, item := range choices {
		// choose next activity based on order result
		var f cadence.Future
		switch item {
		case orderChoiceApple:
			f = cadence.ExecuteActivity(ctx, orderAppleActivity, item)
		case orderChoiceBanana:
			f = cadence.ExecuteActivity(ctx, orderBananaActivity, item)
		case orderChoiceCherry:
			f = cadence.ExecuteActivity(ctx, orderCherryActivity, item)
		case orderChoiceOrange:
			f = cadence.ExecuteActivity(ctx, orderOrangeActivity, item)
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

	cadence.GetActivityLogger(ctx).Info("Get basket order.", zap.Strings("Orders", basket))
	return basket, nil
}
