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

	w.RegisterDynamicWorkflow(dynamic.DynamicWorkflow, workflow.DynamicRegisterOptions{})
	w.RegisterDynamicActivity(dynamic.DynamicActivity, activity.DynamicRegisterOptions{})

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
