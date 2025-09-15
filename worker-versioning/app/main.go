package main

import (
	"context"
	"log"
	"time"

	"github.com/pborman/uuid"

	worker_versioning "github.com/temporalio/samples-go/worker-versioning"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	ctx := context.Background()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// First, we wait for the v1 worker to appear, and we will set it as the current version for
	// the deployment.
	log.Println("Waiting for v1 worker to appear. Run `go run worker-versioning/workerv1/main.go` in another terminal")
	err = waitForWorkerAndMakeCurrent(c, "1.0")
	if err != nil {
		log.Fatalln("Unable to set current deployment version", err)
	}

	// Next we'll start two workflows, one which uses the `AutoUpgrade` behavior, and one which uses
	// `Pinned`.
	autoUpgradeWorkflowId := "worker-versioning-versioning-autoupgrade_" + uuid.New()
	autoUpgradeWorkflowOptions := client.StartWorkflowOptions{
		ID:        autoUpgradeWorkflowId,
		TaskQueue: worker_versioning.TaskQueue,
	}
	autoUpgradeExecution, err := c.ExecuteWorkflow(ctx, autoUpgradeWorkflowOptions, "AutoUpgradingWorkflow")
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	log.Println("Started auto-upgrading workflow",
		"WorkflowID", autoUpgradeExecution.GetID(), "RunID", autoUpgradeExecution.GetRunID())

	pinnedWorkflowId := "worker-versioning-versioning-pinned_" + uuid.New()
	pinnedWorkflowOptions := client.StartWorkflowOptions{
		ID:        pinnedWorkflowId,
		TaskQueue: worker_versioning.TaskQueue,
	}
	pinnedExecution, err := c.ExecuteWorkflow(ctx, pinnedWorkflowOptions, "PinnedWorkflow")
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	log.Println("Started pinned workflow",
		"WorkflowID", pinnedExecution.GetID(), "RunID", pinnedExecution.GetRunID())

	// Signal both workflows a few times to drive them
	err = advanceWorkflows(ctx, c, autoUpgradeExecution, pinnedExecution)
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}

	// Now wait for the v1.1 worker to appear and become current
	log.Println("Waiting for v1.1 worker to appear. Run `go run worker-versioning/workerv1.1/main.go` in another terminal")
	err = waitForWorkerAndMakeCurrent(c, "1.1")
	if err != nil {
		log.Fatalln("Unable to set current deployment version", err)
	}

	// Once it has, we will continue to advance the workflows.
	// The auto-upgrade workflow will now make progress on the new worker, while the pinned one will
	// keep progressing on the old worker.
	err = advanceWorkflows(ctx, c, autoUpgradeExecution, pinnedExecution)
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}

	// Finally we'll start the v2 worker, and again it'll become the new current version
	log.Println("Waiting for v2 worker to appear. Run `go run worker-versioning/workerv2/main.go` in another terminal")
	err = waitForWorkerAndMakeCurrent(c, "2.0")
	if err != nil {
		log.Fatalln("Unable to set current deployment version", err)
	}

	// Once it has we'll start one more new workflow, another pinned one, to demonstrate that new
	// pinned worklfows start on the current version.
	pinnedWorkflow2Id := "worker-versioning-versioning-pinned-2_" + uuid.New()
	pinnedWorkflow2Options := client.StartWorkflowOptions{
		ID:        pinnedWorkflow2Id,
		TaskQueue: worker_versioning.TaskQueue,
	}
	pinnedExecution2, err := c.ExecuteWorkflow(ctx, pinnedWorkflow2Options, "PinnedWorkflow")
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	log.Println("Started pinned workflow v2",
		"WorkflowID", pinnedExecution2.GetID(), "RunID", pinnedExecution2.GetRunID())

	// Now we'll conclude all workflows. You should be able to see in your server UI that the pinned
	// workflow always stayed on 1.0, while the auto-upgrading workflow migrated.
	for _, handle := range []client.WorkflowRun{autoUpgradeExecution, pinnedExecution, pinnedExecution2} {
		err = c.SignalWorkflow(ctx, handle.GetID(), handle.GetRunID(),
			"do-next-signal", "conclude")
		if err != nil {
			log.Fatalln("Unable to signal workflow", err)
			return
		}
		err = handle.Get(ctx, nil)
		if err != nil {
			log.Fatalln("Unable to get workflow result", err)
		}
	}
}

func advanceWorkflows(ctx context.Context, c client.Client, autoUpgradeExecution client.WorkflowRun, pinnedExecution client.WorkflowRun) error {
	var err error
	for i := 0; i < 3; i++ {
		err = c.SignalWorkflow(ctx, autoUpgradeExecution.GetID(), autoUpgradeExecution.GetRunID(),
			"do-next-signal", "do-activity")
		if err != nil {
			return err
		}
		err = c.SignalWorkflow(ctx, pinnedExecution.GetID(), pinnedExecution.GetRunID(),
			"do-next-signal", "some-signal")
		if err != nil {
			return err
		}
	}
	return nil
}

func waitForWorkerAndMakeCurrent(c client.Client, buildID string) error {
	ctx := context.Background()
	deploymentHandle := c.WorkerDeploymentClient().GetHandle(worker_versioning.DeploymentName)
	version := worker.WorkerDeploymentVersion{
		DeploymentName: worker_versioning.DeploymentName,
		BuildId:        buildID,
	}

Outer:
	for {
		d, err := deploymentHandle.Describe(ctx, client.WorkerDeploymentDescribeOptions{})
		if err == nil {
			for _, v := range d.Info.VersionSummaries {
				if v.Version == version {
					break Outer
				}
			}
		}
		time.Sleep(time.Second)
	}

	// Once it has, we will mark this version as the "current" version for the deployment.
	_, err := c.WorkerDeploymentClient().GetHandle(worker_versioning.DeploymentName).SetCurrentVersion(ctx,
		client.WorkerDeploymentSetCurrentVersionOptions{
			BuildID: buildID,
		})
	return err
}
