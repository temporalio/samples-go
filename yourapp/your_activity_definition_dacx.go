package yourapp

import (
	"context"

	"go.temporal.io/sdk/activity"
)

/*
The first parameter of an Activity Definition is `context.Context`.
This parameter is optional for an Activity Definition, though it is recommended, especially if the Activity is expected to use other Go SDK APIs.

An Activity Definition can support as many other custom parameters as needed.
However, all parameters must be serializable (parameters can’t be channels, functions, variadic, or unsafe pointers), and it is recommended to pass a single struct that can be updated later.
*/

// YourActivityParam is the struct passed to your Activity.
// Use a struct so that your function signature remains compatible if fields change.
type YourActivityParam struct {
	ActivityParamX string
	ActivityParamY int
}

/*
A Go-based Activity Definition can return either just an `error` or a `customValue, error` combination (same as a Workflow Definition).
You may wish to use a `struct` type to hold all custom values, just keep in mind they must all be serializable.
*/

// YourActivityResultObject is the struct returned from your Activity.
// Use a struct so that you can return multiple values of different types.
// Additionally, your function signature remains compatible if the fields change.
type YourActivityResultObject struct {
	ResultFieldX string
	ResultFieldY int
}

/*
In the Temporal Go SDK programming model, an Activity Definition is an exportable function or a `struct` method.
Below is an example of a basic exportable function.
*/

// YourSimpleActivityDefinition is a basic Activity Definiton.
func YourSimpleActivityDefinition(ctx context.Context) error {
	return nil
}

/*
Below is an example of an Activity defined as a Struct method.
*/

// YourActivityObject is the struct that maintains shared state across Activities.
// If the Worker crashes this Activity object loses its state.
type YourActivityObject struct {
	SharedMessageState *string
	SharedCounterState *int
}

// YourActivityDefinition is your custom Activity Definition.
// An Activity Definiton is an exportable function.
func (a *YourActivityObject) YourActivityDefinition(ctx context.Context, param YourActivityParam) (YourActivityResultObject, error) {
	// Use Acivities for computations or calling external APIs.
	// This is just an example of appending to text and incrementing a counter.
	message := param.ActivityParamX + " World!"
	counter := param.ActivityParamY + 1
	a.SharedMessageState = &message
	a.SharedCounterState = &counter
	result := YourActivityResultObject{
		ResultFieldX: *a.SharedMessageState,
		ResultFieldY: *a.SharedCounterState,
	}
	// Return the results back to the Workflow Execution.
	// The results persist within the Event History of the Workflow Execution.
	return result, nil
}

/*
An _Activity struct_ can have more than one method, with each method acting as a separate Activity Type.
Activities written as struct methods can use shared struct variables, such as:

- an application level DB pool
- client connection to another service
- reusable utilities
- any other expensive resources that you only want to initialize once per process

Because this is such a common need, the rest of this guide shows Activities written as `struct` methods.
*/

// PrintSharedState is another custom Activity Definition.
func (a *YourActivityObject) PrintSharedSate(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("The current message is:", *a.SharedMessageState)
	logger.Info("The current counter is:", *a.SharedCounterState)
	return nil
}

/* @dac
id: how-to-develop-an-activity-definition-in-go
title: How to develop an Activity Definition in Go
label: Activity Definition
description: In the Temporal Go SDK programming model, an Activity Definition is an exportable function or a `struct` method.
lines: 1-8, 37-60, 74-86
@dac */

/* @dac
id: how-to-define-activity-parameters-in-go
title: How to do define Activity parameters in Go
label: Activity parameters
description: The only required parameter is `context.Context`, but Activities can support many custom parameters.
lines: 9-22, 60, 74
@dac */

/* @dac
id: how-to-define-activity-return-values-in-go
title: How to define Activity return values in Go
label: Activity return values
description: A Go-based Activity Definition can return either just an `error` or a `customValue, error` combination.
lines: 24-35, 60, 67-74
@dac */
