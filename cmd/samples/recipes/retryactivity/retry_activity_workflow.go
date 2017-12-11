package main

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

/**
 * This sample workflow executes unreliable activity and would retry until it reaches a set maximum retry count.
 * It supports custom logic to determine if a retry is needed based on the error. It also support custom back off logic
 * to wait before a retry is issued.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "retryactivityGroup"

// This is registration process where you register all your workflows
// and activity function handlers.
func init() {
	workflow.Register(RetryWorkflow)
	activity.Register(sampleActivity)
}

// RetryWorkflow workflow decider
func RetryWorkflow(ctx workflow.Context, maxRetries int) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// User retry policy.
	backOff := newBackOff(maxRetries)

	err := backOff.Retry(ctx,
		func() (interface{}, error) {
			return nil, workflow.ExecuteActivity(ctx, sampleActivity).Get(ctx, nil)
		})
	if err != nil {
		workflow.GetLogger(ctx).Info("Workflow completed with error.", zap.Error(err))
		return err
	}
	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}

type backOff struct {
	// ...
	// User custom retry policy.
	// This is a simple one.
	// ...
	maxRetries int
}

func newBackOff(maxRetries int) *backOff {
	return &backOff{maxRetries: maxRetries}
}

func (b *backOff) Retry(ctx workflow.Context, op func() (interface{}, error)) error {
	for retryCount := 1; retryCount <= b.maxRetries; retryCount++ {
		_, err := op()

		if err == nil {
			// activity succeed.
			return nil
		}

		// check if we should retry or give up
		if !b.shouldRetry(err) {
			return err
		}

		// optional back off
		workflow.Sleep(ctx, b.backoffDuration(retryCount))
	}
	return errors.New("Exceeded max retry attempts")
}

func (b *backOff) backoffDuration(retryCount int) time.Duration {
	// add custom logic to decide how long to wait before retry, for example exponentially backoff.
	return 0 // 0 indicate to retry immediately
}

func (b *backOff) shouldRetry(err error) bool {
	// add custom logic to decide if we should retry
	switch err.(type) {
	}
	return true
}

/**
 * Unreliable activity that fails randomly
 */
func sampleActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	if rand.Float32() < 0.7 {
		logger.Info("Activity failed, please retry.")
		// Activity could return different error types for different failures so workflow could handle them differently.
		// For example, decide to retry or not based on error type.
		return errors.New("failed")
	}

	logger.Info("Activity succeed.")
	return nil
}
