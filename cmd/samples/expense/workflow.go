package main

import (
	"time"

	"go.uber.org/cadence"
	"go.uber.org/zap"
)

const (
	// ApplicationName is the task list for this sample
	ApplicationName = "expenseGroup"
)

var expenseServerHostPort = "http://localhost:8080"

// This is registration process where you register all your workflow handlers.
func init() {
	cadence.RegisterWorkflow(SampleExpenseWorkflow)
}

// SampleExpenseWorkflow workflow decider
func SampleExpenseWorkflow(ctx cadence.Context, expenseID string) (result string, err error) {
	// step 1, create new expense report
	ao := cadence.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx1 := cadence.WithActivityOptions(ctx, ao)
	logger := cadence.GetLogger(ctx)

	err = cadence.ExecuteActivity(ctx1, createExpenseActivity, expenseID).Get(ctx1, nil)
	if err != nil {
		logger.Error("Failed to create expense report", zap.Error(err))
		return "", err
	}

	// step 2, wait for the expense report to be approved (or rejected)
	ao = cadence.ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Minute,
		StartToCloseTimeout:    10 * time.Minute,
	}
	ctx2 := cadence.WithActivityOptions(ctx, ao)
	// Notice that we set the timeout to be 10 minutes for this sample demo. If the expected time for the activity to
	// complete (waiting for human to approve the request) is longer, you should set the timeout accordingly so the
	// cadence system will wait accordingly. Otherwise, cadence system could mark the activity as failure by timeout.
	var status string
	err = cadence.ExecuteActivity(ctx2, waitForDecisionActivity, expenseID).Get(ctx2, &status)
	if err != nil {
		return "", err
	}

	if status != "APPROVED" {
		logger.Info("Workflow completed.", zap.String("ExpenseStatus", status))
		return "", nil
	}

	// step 3, request payment to the expense
	err = cadence.ExecuteActivity(ctx2, paymentActivity, expenseID).Get(ctx2, nil)
	if err != nil {
		logger.Info("Workflow completed with payment failed.", zap.Error(err))
		return "", err
	}

	logger.Info("Workflow completed with expense payment completed.")
	return "COMPLETED", nil
}
