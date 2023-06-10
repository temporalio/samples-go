package main

import (
	"github.com/pborman/uuid"
	build_id_versioning "github.com/temporalio/samples-go/build-id-versioning"
	"sync"

	"log"

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

	taskQueue := "build-id-versioning-" + uuid.New()
	log.Println("Using Task Queue name: ", taskQueue, "(Copy this!)")

	// We will start a handful of workers, each with a different build identifier.
	wg := sync.WaitGroup{}
	createAndRunWorker(c, taskQueue, "1.0", build_id_versioning.SampleChangingWorkflowV1, &wg)
	createAndRunWorker(c, taskQueue, "1.1", build_id_versioning.SampleChangingWorkflowV1b, &wg)
	createAndRunWorker(c, taskQueue, "2.0", build_id_versioning.SampleChangingWorkflowV2, &wg)
	wg.Wait()
}

func createAndRunWorker(c client.Client, taskQueue, buildID string, workflowFunc func(ctx workflow.Context) error, wg *sync.WaitGroup) {
	w := worker.New(c, taskQueue, worker.Options{
		// Both of these options must be set to opt into the feature
		BuildID:                 buildID,
		UseBuildIDForVersioning: true,
	})
	// It's important that we register all the different implementations of the workflow using
	// the same name. This allows us to demonstrate what would happen if you were making changes
	// to this workflow code over time while keeping the same workflow name/type.
	w.RegisterWorkflowWithOptions(workflowFunc, workflow.RegisterOptions{Name: "SampleChangingWorkflow"}) //workflowcheck:ignore
	w.RegisterActivity(build_id_versioning.SomeActivity)
	w.RegisterActivity(build_id_versioning.SomeIncompatibleActivity)

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := w.Run(worker.InterruptCh())
		if err != nil {
			log.Fatalf("Unable to start %s worker: %v", buildID, err)
		}
	}()
}
