package dynamic_workflows

import (
	"context"
	"fmt"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/converter"
)

func DynamicActivity(ctx context.Context, args converter.EncodedValues) (string, error) {
	var arg1, arg2 string
	err := args.Get(&arg1, &arg2)
	if err != nil {
		return "", fmt.Errorf("failed to decode arguments: %w", err)
	}

	info := activity.GetInfo(ctx)
	result := fmt.Sprintf("%s - %s - %s", info.WorkflowType.Name, arg1, arg2)

	return result, nil
}
