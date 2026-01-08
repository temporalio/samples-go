package dataconverterlargepayloads

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// Custom Data to send to temporal.
type CustomData struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
}

// Hold a lot of data.
type ValueContainer struct {
	Values []CustomData `json:"custom_data"`
}

func Workflow(ctx workflow.Context) error {

	logger := workflow.GetLogger(ctx)

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	small := make([]CustomData, 0, 100)
	for i := range 100 {
		small = append(small, CustomData{
			Index: i,
			Name:  fmt.Sprintf("custom_data_%7d", i),
		})
	}
	var retsmall *ValueContainer
	err := workflow.ExecuteActivity(ctx, Activity, &ValueContainer{
		Values: small,
	}).Get(ctx, &retsmall)
	if err != nil {
		return fmt.Errorf("failed running small: %w", err)
	}
	logger.Info("Ran small", "first", retsmall.Values[0], "last", retsmall.Values[len(retsmall.Values)-1])

	huge := make([]CustomData, 0, 100_000)
	for i := range 100_000 {
		huge = append(huge, CustomData{
			Index: i,
			Name:  fmt.Sprintf("custom_data_%7d", i),
		})
	}
	var rethuge *ValueContainer
	err = workflow.ExecuteActivity(ctx, Activity, &ValueContainer{
		Values: huge,
	}).Get(ctx, &rethuge)
	if err != nil {
		return fmt.Errorf("failed running huge: %w", err)
	}
	logger.Info("Ran huge", "first", rethuge.Values[0], "last", rethuge.Values[len(rethuge.Values)-1])

	return nil
}

func Activity(ctx context.Context, input *ValueContainer) (*ValueContainer, error) {
	logger := activity.GetLogger(ctx)

	logger.Info("Got input", "total_values", len(input.Values))
	logger.Info("Activity completed")

	return input, nil
}
