package main

import (
	"log"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"samples-go/yourapp"
)

/*
Create an instance of [`Worker`](https://pkg.go.dev/go.temporal.io/sdk/worker#Worker) by calling [`worker.New()`](https://pkg.go.dev/go.temporal.io/sdk/worker#New), available through the `go.temporal.io/sdk/worker` package, and pass it the following parameters:

1. An instance of the Temporal Go SDK `Client`.
1. The name of the Task Queue that it will poll.
1. An instance of `worker.Options`, which can be empty.

Then, register the Workflow Types and the Activity Types that the Worker will be capable of executing.

Lastly, call either the `Start()` or the `Run()` method on the instance of the Worker.
Run accepts an interrupt channel as a parameter, so that the Worker can be stopped in the terminal.
Otherwise, the `Stop()` method must be called to stop the Worker.
*/

func main() {
	// Create a Temporal Client
	// A Temporal Client is a heavyweight object that should be created just once per process.
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer temporalClient.Close()
	// Create a new Worker.
	yourWorker := worker.New(temporalClient, "your-custom-task-queue-name", worker.Options{})
	// Register your Workflow Definitions with the Worker.
	// Use the ReisterWorkflow or RegisterWorkflowWithOptions method for each Workflow registration.
	yourWorker.RegisterWorkflow(yourapp.YourWorkflowDefinition)
	// Use RegisterOptions to set the name of the Workflow Type for example.
	registerWFOptions := workflow.RegisterOptions{
		Name: "JustAnotherWorkflow",
	}
	yourWorker.RegisterWorkflowWithOptions(yourapp.YourSimpleWorkflowDefinition, registerWFOptions)
	// Register your Activity Definitons with the Worker.
	// Use this technique for registering all Activities that are part of a struct and set the shared variable values.
	initialMessageString := "No messages!"
	initialCounterState := 0
	activities := &yourapp.YourActivityObject{
		SharedMessageState: &initialMessageString,
		SharedCounterState: &initialCounterState,
	}
	// Use the RegisterActivity or RegisterActivityWithOptions method for each Activity.
	yourWorker.RegisterActivity(activities)
	// Use RegisterOptions to change the name of the Activity Type for example.
	registerAOptions := activity.RegisterOptions{
		Name: "JustAnotherActivity",
	}
	yourWorker.RegisterActivityWithOptions(yourapp.YourSimpleActivityDefinition, registerAOptions)
	// Run the Worker
	err = yourWorker.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start Worker", err)
	}
}

/*
In Go, by default, the Workflow Type name is the same as the function name.

To customize the Workflow Type, set the `Name` parameter with `RegisterOptions` when registering your Workflow with a Worker.
*/

/*

:::tip

If you have [`gow`](https://github.com/mitranim/gow) installed, the Worker Process automatically "reloads" when you update the Worker file:

```bash
go install github.com/mitranim/gow@latest
gow run worker/main.go # automatically reloads when file changes
```

:::

*/

/* @dac
id: how-to-develop-a-worker-in-go
title: How to develop a Worker in Go
label: Develop Worker
description: Develop an instance of a Worker by calling worker.New(), available via the go.temporal.io/sdk/worker package.
lines: 1-40, 46-55, 61-66, 74-87
@dac */

/* @dac
id: how-to-customize-workflow-type-in-go
title: How to customize Workflow Type in Go
label: Customize Workflow Type
description: To customize the Workflow Type set the Name parameter with RegisterOptions when registering your Workflow with a Worker.
lines: 1-12, 28, 37, 41-45, 66-72
@dac */
