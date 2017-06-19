package main

import (
	"errors"
	"fmt"

	"go.uber.org/cadence"
	"go.uber.org/zap"
)

/**
 * This sample workflow demonstrates how to use invoke child workflow from parent workflow execution.  Each child
 * workflow execution is starting a new run and parent execution is notified only after the completion of last run.
 */

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	cadence.RegisterWorkflow(SampleChildWorkflow)
}

// SampleChildWorkflow workflow decider
func SampleChildWorkflow(ctx cadence.Context, totalCount, runCount int) (string, error) {
	logger := cadence.GetLogger(ctx)
	logger.Info("Child workflow execution started.")
	if runCount <= 0 {
		logger.Error("Invalid valid for run count.", zap.Int("RunCount", runCount))
		return "", errors.New("invalid run count")
	}

	totalCount++
	runCount--
	if runCount == 0 {
		result := fmt.Sprintf("Child workflow execution completed after %v runs", totalCount)
		logger.Info("Child workflow completed.", zap.String("Result", result))
		return result, nil
	}

	logger.Info("Child workflow starting new run.", zap.Int("RunCount", runCount), zap.Int("TotalCount",
		totalCount))
	return "", cadence.NewContinueAsNewError(ctx, SampleChildWorkflow, totalCount, runCount)
}
