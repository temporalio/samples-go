package main

import (
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	// ApplicationName is the task list for this sample
	ApplicationName = "expenseGroup"
)

var expenseServerHostPort = "http://localhost:8099"

// This is registration process where you register all your workflow handlers.
func init() {
	workflow.Register(SampleExpenseWorkflow)
}

// SampleExpenseWorkflow workflow decider
func SampleExpenseWorkflow(ctx workflow.Context, expenseID string) (result string, err error) {
	// step 1, create new expense report
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)

	err = workflow.ExecuteActivity(ctx1, createExpenseActivity, expenseID).Get(ctx1, nil)
	if err != nil {
		logger.Error("Failed to create expense report", zap.Error(err))
		return "", err
	}

	// step 2, wait for the expense report to be approved (or rejected)
	ao = workflow.ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Minute,
		StartToCloseTimeout:    10 * time.Minute,
	}
	ctx2 := workflow.WithActivityOptions(ctx, ao)
	// Notice that we set the timeout to be 10 minutes for this sample demo. If the expected time for the activity to
	// complete (waiting for human to approve the request) is longer, you should set the timeout accordingly so the
	// cadence system will wait accordingly. Otherwise, cadence system could mark the activity as failure by timeout.
	var status string
	err = workflow.ExecuteActivity(ctx2, waitForDecisionActivity, expenseID).Get(ctx2, &status)
	if err != nil {
		return "", err
	}

	if status != "APPROVED" {
		logger.Info("Workflow completed.", zap.String("ExpenseStatus", status))
		return "", nil
	}

	// step 3, request payment to the expense
	err = workflow.ExecuteActivity(ctx2, paymentActivity, expenseID).Get(ctx2, nil)
	if err != nil {
		logger.Info("Workflow completed with payment failed.", zap.Error(err))
		return "", err
	}

	logger.Info("Workflow completed with expense payment completed.")
	return "COMPLETED", nil
}
