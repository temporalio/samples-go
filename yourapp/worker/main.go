// @@@SNIPSTART go-samples-yourapp-your-worker
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/yourapp"
)

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
	yourWorker.RegisterWorkflow(yourapp.YourWorkflowDefinition)
	// Use the ReisterWorkflow method for each function registration.
	yourWorker.RegisterWorkflow(yourapp.YourSimpleWorkflowDefinition)
	// Register your Activity Definitons with the Worker.
	// Use this technique for registering all Activities that are part of a struct and set the shared variable values.
	initialMessageString := "No messages!"
	initialCounterState := 0
	activities := &yourapp.YourActivityObject{
		SharedMessageState: &initialMessageString,
		SharedCounterState: &initialCounterState,
	}
	yourWorker.RegisterActivity(activities)
	yourWorker.RegisterActivity(yourapp.YourSimpleActivityDefinition)
	// Run the Worker
	err = yourWorker.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start Worker", err)
	}
}
// @@@SNIPEND
