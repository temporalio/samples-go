package workflowactivity

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	temporalotel "go.temporal.io/sdk/contrib/opentelemetry-v2"
	"go.temporal.io/sdk/workflow"
)

// @@@SNIPSTART samples-go-opentelemetry-v2-application-spans
const instrumentationName = "github.com/temporalio/samples-go/opentelemetry-v2/workflow-activity-propagation"

func Workflow(ctx workflow.Context, name string) (string, error) {
	tracer := temporalotel.Tracer(instrumentationName)
	ctx, span := tracer.Start(ctx, "workflow-operation")
	defer span.End()

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	})

	var result string
	if err := workflow.ExecuteActivity(ctx, Activity, name).Get(ctx, &result); err != nil {
		return "", err
	}

	return result, nil
}

func Activity(ctx context.Context, name string) (string, error) {
	_, span := otel.Tracer(instrumentationName).Start(ctx, "activity-operation")
	defer span.End()

	return fmt.Sprintf("Hello, %s!", name), nil
}

// @@@SNIPEND
