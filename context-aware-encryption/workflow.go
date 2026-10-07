package contextawareencryption

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
)

// TenantWorkflow is a standard workflow definition.
// Note that the TenantWorkflow and TenantActivity don't need to care that
// their inputs/results are being encrypted/decrypted.
func TenantWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	//logger.Info("Encrypted Payloads workflow started", "name", name)
	value, ok := workflow.Context.Value(ctx, PropagateKey).(CryptContext)
	if !ok {
		logger.Error("Unable to retrieve context")
	}
	logger.Info("Context KeyID", value)
	info := map[string]string{
		"name": name,
	}

	var result string
	err := workflow.ExecuteActivity(ctx, TenantActivity, info).Get(ctx, &result)
	if err != nil {
		logger.Error("TenantActivity failed.", "Error", err)
		return "", err
	}

	//logger.Info("TenantWorkflow.", "result", result)

	return result, nil
}

func TenantActivity(ctx context.Context, info map[string]string) (string, error) {
	logger := activity.GetLogger(ctx)
	value, ok := context.Context.Value(ctx, PropagateKey).(CryptContext)
	if !ok {
		logger.Error("Activity Unable to retrieve context")
		return "", fmt.Errorf("Unable to retrieve context")
	}
	fmt.Println("Activity Context Value:", value)
	//logger.Info("TenantActivity", "info", info)

	name, ok := info["name"]
	if !ok {
		name = "someone"
	}

	return "Hello " + name + "!", nil
}
