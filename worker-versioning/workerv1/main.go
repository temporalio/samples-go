package main

import (
	"log"

	worker_versioning "github.com/temporalio/samples-go/worker-versioning"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, worker_versioning.TaskQueue, worker.Options{
		DeploymentOptions: worker.DeploymentOptions{
			UseVersioning: true,
			Version: worker.WorkerDeploymentVersion{
				DeploymentName: worker_versioning.DeploymentName,
				BuildId:        "1.0",
			},
		},
	})
	// It's important that we register all the different implementations of the workflow using
	// the same name. This allows us to demonstrate what would happen if you were making changes
	// to this workflow code over time while keeping the same workflow name/type.
	w.RegisterWorkflowWithOptions(worker_versioning.AutoUpgradingWorkflowV1, workflow.RegisterOptions{
		Name:               "AutoUpgradingWorkflow",
		VersioningBehavior: workflow.VersioningBehaviorAutoUpgrade,
	})
	w.RegisterWorkflowWithOptions(worker_versioning.PinnedWorkflowV1, workflow.RegisterOptions{
		Name:               "PinnedWorkflow",
		VersioningBehavior: workflow.VersioningBehaviorPinned,
	})
	w.RegisterActivity(worker_versioning.SomeActivity)
	w.RegisterActivity(worker_versioning.SomeIncompatibleActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Unable to start worker: %v", err)
	}
}
