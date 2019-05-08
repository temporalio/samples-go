package main

import (
	"context"
	"fmt"

	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

/**
 * Sample activities used by file processing sample workflow.
 */
const (
	evaluateFitnessActivityName = "evaluateFitnessActivity"
)

// This is registration process where you register all your activity handlers.
func init() {
	activity.RegisterWithOptions(
		evaluateFitnessActivity,
		activity.RegisterOptions{Name: evaluateFitnessActivityName},
	)
}

func evaluateFitnessActivity(ctx context.Context, functionName string, location []float64) (float64, error) {
	logger := activity.GetLogger(ctx).With(zap.String("HostID", HostID))
	logger.Info("EvaluateFitnessActivity started.")

	// i := 0
	// if activity.HasHeartbeatDetails(ctx) {
	// 	// we are retry from a failed attempt, and there is reported progress that we should resume from.
	// 	var completedIdx int
	// 	if err := activity.GetHeartbeatDetails(ctx, &completedIdx); err == nil {
	// 		i = completedIdx + 1
	// 		logger.Info("Resuming from failed attempt", zap.Int("ReportedProgress", completedIdx))
	// 	}
	// }

	// for ; i < 10; i++ {
	// 	// process task i
	// 	logger.Info("processing task", zap.Int("TaskID", i))
	// 	activity.RecordHeartbeat(ctx, i)

	// 	// simulate failure after process 1/3 of the tasks
	// 	if failed {
	// 		logger.Info("Activity failed, will retry...")
	// 		// Activity could return different error types for different failures so workflow could handle them differently.
	// 		// For example, decide to retry or not based on error reasons.
	// 		return cadence.NewCustomError("some-retryable-error")
	// 	}
	// }

	var function ObjectiveFunction

	switch functionName {
	case "sphere":
		function = Sphere
	case "rosenbrock":
		function = Rosenbrock
	case "griewank":
		function = Griewank
	}

	value := function.Evaluate(location)

	logger.Info(fmt.Sprintf("Activity succeed, fitness=%f", value))

	return value, nil
}
