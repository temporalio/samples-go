// @@@SNIPSTART go-samples-your-activity-definition
package yourapp

import (
	"context"

	"go.temporal.io/sdk/activity"
)

// Use a struct so that your function signature remains compatible if fields change.
type YourActivityParam struct {
	ActivityParamX string
	ActivityParamY int
}

// Use a struct so that you can return multiple values of different types.
// Additionally, your function signature remains compatible if the fields change.
type YourActivityResultObject struct {
	ResultFieldX string
	ResultFieldY int
}

// If the Worker crashes this Activity object loses its state.
type YourActivityObject struct {
	SharedMessageState string
	SharedCounterState int
}

// An Activity Definiton is an exportable function.
func (a *YourActivityObject) YourActivityDefinition(ctx context.Context, param YourActivityParam) (YourActivityResultObject, error) {
	// Use Acivities for computations or calling external APIs.
	// This is just an example of appending to text and incrementing a counter.
	a.SharedMessageState = param.ActivityParamX + " World!"
	a.SharedCounterState = param.ActivityParamY + 1
	result := YourActivityResultObject{
		ResultFieldX: a.SharedMessageState,
		ResultFieldY: a.SharedCounterState,
	}
	// Return the results back to the Workflow Execution.
	// The results persist within the Event History of the Workflow Execution.
	return result, nil
}

func (a *YourActivityObject) PrintSharedSate(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("The current message is:", a.SharedMessageState)
	logger.Info("The current counter is:", a.SharedCounterState)
	return nil
}

// An Activity Definiton is an exportable function.
func YourSimpleActivityDefinition(ctx context.Context) error {
	return nil
}

// @@@SNIPEND
