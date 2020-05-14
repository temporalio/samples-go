package choice

import (
	"time"

	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
)

const (
	OrderChoiceApple  = "apple"
	OrderChoiceBanana = "banana"
	OrderChoiceCherry = "cherry"
	OrderChoiceOrange = "orange"
)

// ExclusiveChoiceWorkflow Workflow definition.
func ExclusiveChoiceWorkflow(ctx workflow.Context) error {
	// Get order.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
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
		logger.Error("Unexpected order", zap.String("Choice", orderChoice))
	}

	logger.Info("Workflow completed.")
	return nil
}
