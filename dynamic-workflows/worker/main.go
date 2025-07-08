package main

import (
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/workflow"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	dynamic "github.com/temporalio/samples-go/dynamic-workflows"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "dynamic-workflows", worker.Options{})

	// Dynamic runtimes options are specified at workflow registration time.
	options := workflow.DynamicRegisterOptions{
		LoadDynamicRuntimeOptions: func(details workflow.LoadDynamicRuntimeOptionsDetails) (workflow.DynamicRuntimeOptions, error) {
			var options workflow.DynamicRuntimeOptions
			switch details.WorkflowType.Name {
			case "some-workflow":
				options.VersioningBehavior = workflow.VersioningBehaviorAutoUpgrade
			case "behavior-pinned":
				options.VersioningBehavior = workflow.VersioningBehaviorPinned
			default:
				options.VersioningBehavior = workflow.VersioningBehaviorUnspecified
			}
			return options, nil
		},
	}

	w.RegisterDynamicWorkflow(dynamic.DynamicWorkflow, options)
	w.RegisterDynamicActivity(dynamic.DynamicActivity, activity.DynamicRegisterOptions{})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
