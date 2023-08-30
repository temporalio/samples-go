package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.temporal.io/api/workflowservice/v1"

	"github.com/pborman/uuid"

	"go.temporal.io/sdk/client"
)

func main() {
	ctx := context.Background()

	// Get task queue name from CLI arg
	taskQueue := os.Args[1]
	if taskQueue == "" {
		log.Fatalln("Must provide task queue name as first and only argument")
	}
	log.Println("Using task queue", taskQueue)

	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// First, let's make the task queue use the build id versioning feature by adding an initial
	// default version to the queue:
	err = c.UpdateWorkerBuildIdCompatibility(ctx, &client.UpdateWorkerBuildIdCompatibilityOptions{
		TaskQueue: taskQueue,
		Operation: &client.BuildIDOpAddNewIDInNewDefaultSet{
			BuildID: "1.0",
		},
	})
	if err != nil {
		log.Fatalln("Unable to update worker build id compatibility", err)
	}

	firstWorkflowID := "build-id-versioning-first_" + uuid.New()
	firstWorkflowOptions := client.StartWorkflowOptions{
		ID:                       firstWorkflowID,
		TaskQueue:                taskQueue,
		WorkflowExecutionTimeout: 5 * time.Minute,
	}
	firstExecution, err := c.ExecuteWorkflow(ctx, firstWorkflowOptions, "SampleChangingWorkflow")
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	log.Println("Started first workflow",
		"WorkflowID", firstExecution.GetID(), "RunID", firstExecution.GetRunID())

	// Signal this workflow a few times to drive it
	for i := 0; i < 3; i++ {
		err = c.SignalWorkflow(ctx, firstExecution.GetID(), firstExecution.GetRunID(),
			"do-next-signal", "do-activity")
		if err != nil {
			log.Fatalln("Unable to signal workflow", err)
		}
	}

	// Give a chance for these signals to be processed by the 1.0 worker
	time.Sleep(5 * time.Second)

	// Now, let's update the task queue with a new compatible version:
	err = c.UpdateWorkerBuildIdCompatibility(ctx, &client.UpdateWorkerBuildIdCompatibilityOptions{
		TaskQueue: taskQueue,
		Operation: &client.BuildIDOpAddNewCompatibleVersion{
			BuildID:                   "1.1",
			ExistingCompatibleBuildID: "1.0",
		},
	})
	if err != nil {
		log.Fatalln("Unable to update build id compatability", err)
	}

	// Continue driving the workflow. Take note that the new version of the workflow run by the
	// 1.1 worker is the one that takes over! You might see a workflow task timeout, if the 1.0
	// worker is processing a task as the version update happens. That's normal.
	for i := 0; i < 3; i++ {
		err = c.SignalWorkflow(ctx, firstExecution.GetID(), firstExecution.GetRunID(),
			"do-next-signal", "do-activity")
		if err != nil {
			log.Fatalln("Unable to signal workflow", err)
		}
	}

	// Add a new *incompatible* version to the task queue, which will become the new overall default
	// for the queue.
	err = c.UpdateWorkerBuildIdCompatibility(ctx, &client.UpdateWorkerBuildIdCompatibilityOptions{
		TaskQueue: taskQueue,
		Operation: &client.BuildIDOpAddNewIDInNewDefaultSet{
			BuildID: "2.0",
		},
	})
	if err != nil {
		log.Fatalln("Unable to update build id compatability", err)
	}

	// Start a new workflow, note that it will run on the new 2.0 version, without the client
	// invocation changing at all!
	secondWorkflowID := "build-id-versioning-second_" + uuid.New()
	secondWorkflowOptions := client.StartWorkflowOptions{
		ID:                       secondWorkflowID,
		TaskQueue:                taskQueue,
		WorkflowExecutionTimeout: 5 * time.Minute,
	}
	secondExecution, err := c.ExecuteWorkflow(ctx, secondWorkflowOptions, "SampleChangingWorkflow")
	if err != nil {
		log.Fatalln("Unable to start workflow", err)
	}
	log.Println("Started second workflow",
		"WorkflowID", secondExecution.GetID(), "RunID", secondExecution.GetRunID())

	// Drive the first workflow to completion, the second will finish on its own
	err = c.SignalWorkflow(ctx, firstExecution.GetID(), firstExecution.GetRunID(),
		"do-next-signal", "do-activity")
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}
	err = c.SignalWorkflow(ctx, firstExecution.GetID(), firstExecution.GetRunID(),
		"do-next-signal", "finish")
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}

	// Lastly we'll demonstrate how you can use the gRPC api to determine if certain bulid IDs are
	// ready to be retied. There's more information in the documentation, but here's a quick example
	// that will show us that we can retire the 1.0 worker:
	retirementInfo, err := c.WorkflowService().GetWorkerTaskReachability(ctx, &workflowservice.GetWorkerTaskReachabilityRequest{
		Namespace: "default",
		BuildIds:  []string{"1.0"},
	})
	if err != nil {
		log.Fatalln("Unable to get build id reachability", err)
	}
	reachabilityOf1Dot0 := retirementInfo.GetBuildIdReachability()[0]
	noReachableQueues := true
	for _, tq := range reachabilityOf1Dot0.GetTaskQueueReachability() {
		if tq.GetReachability() != nil && len(tq.GetReachability()) > 0 {
			noReachableQueues = false
		}
	}
	if noReachableQueues {
		log.Println("We have determined 1.0 is ready to be retired")
	}

}
