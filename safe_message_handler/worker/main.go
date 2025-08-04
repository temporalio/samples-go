package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/safe_message_handler"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "safe-message-handlers-task-queue", worker.Options{})

	w.RegisterWorkflow(safe_message_handler.ClusterManagerWorkflow)
	w.RegisterActivity(safe_message_handler.AssignNodesToJobsActivity)
	w.RegisterActivity(safe_message_handler.UnassignNodesForJobActivity)
	w.RegisterActivity(safe_message_handler.FindBadNodesActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
