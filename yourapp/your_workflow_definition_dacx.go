package yourapp

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

/*
The first parameter of a Go-based Workflow Definition must be of the [`workflow.Context`](https://pkg.go.dev/go.temporal.io/sdk/workflow#Context) type.
It is used by the Temporal Go SDK to pass around Workflow Execution context, and virtually all the Go SDK APIs that are callable from the Workflow require it.
It is acquired from the [`go.temporal.io/sdk/workflow`](https://pkg.go.dev/go.temporal.io/sdk/workflow) package.

The `workflow.Context` entity operates similarly to the standard `context.Context` entity provided by Go.
The only difference between `workflow.Context` and `context.Context` is that the `Done()` function, provided by `workflow.Context`, returns `workflow.Channel` instead of the standard Go `chan`.

The second parameter, `string`, is a custom parameter that is passed to the Workflow when it is invoked.
A Workflow Definition may support multiple custom parameters, or none.
These parameters can be regular type variables or safe pointers.
However, the best practice is to pass a single parameter that is of a `struct` type, so there can be some backward compatibility if new parameters are added.

All Workflow Definition parameters must be serializable and can't be channels, functions, variadic, or unsafe pointers.
*/

// YourWorkflowParam is the object passed to the Workflow.
type YourWorkflowParam struct {
	WorkflowParamX string
	WorkflowParamY int
}

/*
A Go-based Workflow Definition can return either just an `error` or a `customValue, error` combination.
Again, the best practice here is to use a `struct` type to hold all custom values.
*/

// YourWorkflowResultObject is the object returned by the Workflow.
type YourWorkflowResultObject struct {
	WFResultFieldX string
	WFResultFieldY int
}

/*
In the Temporal Go SDK programming model, a [Workflow Definition](/concepts/what-is-a-workflow-definition) is an exportable function.
Below is an example of a basic Workflow Definition.
*/

// YourSimpleWorkflowDefintiion is the most basic Workflow Defintion.
func YourSimpleWorkflowDefinition(ctx workflow.Context) error {
	// ...
	return nil
}

// YourWorkflowDefinition is your custom Workflow Definition.
func YourWorkflowDefinition(ctx workflow.Context, param YourWorkflowParam) (*YourWorkflowResultObject, error) {
	// Set the options for the Activity Execution.
	// Either StartToClose Timeout OR ScheduleToClose is required.
	// Not specifying a Task Queue will default to the parent Workflow Task Queue.
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)
	activityParam := YourActivityParam{
		ActivityParamX: param.WorkflowParamX,
		ActivityParamY: param.WorkflowParamY,
	}
	// Use a nil struct pointer to call Activities that are part of a struct.
	var a *YourActivityObject
	err := workflow.ExecuteActivity(ctx, a.PrintSharedSate).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	// Execute the Activity and wait for the result.
	var activityResult YourActivityResultObject
	err = workflow.ExecuteActivity(ctx, a.YourActivityDefinition, activityParam).Get(ctx, &activityResult)
	if err != nil {
		return nil, err
	}
	// Execute another Activity and wait for the result.
	err = workflow.ExecuteActivity(ctx, a.PrintSharedSate).Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	// Make the results of the Workflow Execution available.
	workflowResult := &YourWorkflowResultObject{
		WFResultFieldX: activityResult.ResultFieldX,
		WFResultFieldY: activityResult.ResultFieldY,
	}
	return workflowResult, nil
}

/*
A Workflow Definition written in Go can return both a custom value and an error.
However, it's not possible to receive both a custom value and an error in the calling process, as is normal in Go.
The caller will receive either one or the other.
Returning a non-nil `error` from a Workflow indicates that an error was encountered during its execution and the Workflow Execution should be terminated, and any custom return values will be ignored by the system.
*/

/*
In Go, Workflow Definition code cannot directly do the following:

- Iterate over maps using `range`, because with `range` the order of the map's iteration is randomized.
  Instead you can collect the keys of the map, sort them, and then iterate over the sorted keys to access the map.
  This technique provides deterministic results.
  You can also use a Side Effect or an Activity to process the map instead.
- Call an external API, conduct a file I/O operation, talk to another service, etc. (Use an Activity for these.)

The Temporal Go SDK has APIs to handle equivalent Go constructs:

- `workflow.Now()` This is a replacement for `time.Now()`.
- `workflow.Sleep()` This is a replacement for `time.Sleep()`.
- `workflow.GetLogger()` This ensures that the provided logger does not duplicate logs during a replay.
- `workflow.Go()` This is a replacement for the `go` statement.
- `workflow.Channel` This is a replacement for the native `chan` type.
  Temporal provides support for both buffered and unbuffered channels.
- `workflow.Selector` This is a replacement for the `select` statement.
  Learn more on the [Go SDK Selectors](https://legacy-documentation-sdks.temporal.io/go/selectors) page.
- `workflow.Context` This is a replacement for `context.Context`.
  See [Tracing](/app-dev-context/tracing) for more information about context propagation.
*/

/* @dac
id: how-to-develop-a-workflow-definition-in-go
title: How to develop a Workflow Definition in Go
label: Workflow Definition
description: In the Temporal Go SDK programming model, a Workflow Definition is an exportable function.
lines: 1-7, 47-51
@dac */

/* @dac
id: how-to-define-workflow-parameters-in-go
title: How to define Workflow parameters in Go
label: Workflow parameters
description: A Go-based Workflow Definition must accept workflow.Context and may support multiple custom parameters.
lines:  1-29, 53-54, 89
@dac */

/* @dac
id: how-to-define-workflow-return-values-in-go
title: How to define Workflow return values in Go
label: Workflow return values
description: A Go-based Workflow Definition can return either just an `error` or a `customValue, error` combination.
lines: 1-7, 31-40, 53-54, 83-96
@dac */

/*dac
id: how-to-handle-workflow-logic-requirements-in-go
title: How to handle Workflow logic requirements in Go
label: Workflow logic requirements
description: In Go, Workflow Definition code cannot directly do a few things to adhere to deterministic constraints.
lnes: 98-119
@dac */
