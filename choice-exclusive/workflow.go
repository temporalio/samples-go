package choice

import (
	"fmt"
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
		err = workflow.ExecuteActivity(ctx, orderActivities.OrderApple, orderChoice).Get(ctx, nil)
	case OrderChoiceBanana:
		err = workflow.ExecuteActivity(ctx, orderActivities.OrderBanana, orderChoice).Get(ctx, nil)
	case OrderChoiceCherry:
		err = workflow.ExecuteActivity(ctx, orderActivities.OrderCherry, orderChoice).Get(ctx, nil)
	case OrderChoiceOrange:
		// Activity can be also called by its name
		err = workflow.ExecuteActivity(ctx, "OrderOrange", orderChoice).Get(ctx, nil)
	default:
		err = fmt.Errorf("unknown order choice: %v", orderChoice)
	}

	if err != nil {
		return err
	}

	logger.Info("Workflow completed.")
	return nil
}
