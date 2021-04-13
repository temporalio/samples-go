package choice

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

const (
	OrderChoiceApple  = "apple"
	OrderChoiceBanana = "banana"
	OrderChoiceCherry = "cherry"
	OrderChoiceOrange = "orange"
)

// ExclusiveChoiceWorkflow Workflow definition.
func ExclusiveChoiceWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Get order.
	var orderActivities *OrderActivities // Used to call activities by function pointer
	var orderChoice string
	err := workflow.ExecuteActivity(ctx, orderActivities.GetOrder).Get(ctx, &orderChoice)
	if err != nil {
		return err
	}

	logger := workflow.GetLogger(ctx)

	// choose next activity based on order result
	switch orderChoice {
	case OrderChoiceApple:
		workflow.ExecuteActivity(ctx, orderActivities.OrderApple, orderChoice)
	case OrderChoiceBanana:
		workflow.ExecuteActivity(ctx, orderActivities.OrderBanana, orderChoice)
	case OrderChoiceCherry:
		workflow.ExecuteActivity(ctx, orderActivities.OrderCherry, orderChoice)
	case OrderChoiceOrange:
		// Activity can be also called by its name
		workflow.ExecuteActivity(ctx, "OrderOrange", orderChoice)
	default:
		logger.Error("Unexpected order", "Choice", orderChoice)
	}

	logger.Info("Workflow completed.")
	return nil
}
