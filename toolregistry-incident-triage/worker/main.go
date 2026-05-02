// Worker entrypoint for the Go triage worker.
package main

import (
	"log"
	"os"

	"github.com/temporalio/samples-go/toolregistry-incident-triage/activities"
	"github.com/temporalio/samples-go/toolregistry-incident-triage/workflows"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	address := os.Getenv("TEMPORAL_ADDRESS")
	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	apiKey := os.Getenv("TEMPORAL_API_KEY")
	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		taskQueue = "triage-go"
	}
	if address == "" || namespace == "" || apiKey == "" {
		log.Fatalln("missing TEMPORAL_ADDRESS / TEMPORAL_NAMESPACE / TEMPORAL_API_KEY")
	}

	log.Printf("connecting to %s (ns=%s) on task queue %s", address, namespace, taskQueue)

	c, err := client.Dial(client.Options{
		HostPort:    address,
		Namespace:   namespace,
		Credentials: client.NewAPIKeyStaticCredentials(apiKey),
	})
	if err != nil {
		log.Fatalf("temporal dial: %v", err)
	}
	defer c.Close()

	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(workflows.IncidentTriageWorkflow)
	w.RegisterWorkflow(workflows.ApprovalWorkflow)
	w.RegisterActivityWithOptions(activities.TriageIncidentActivity, activity.RegisterOptions{
		Name: workflows.TriageActivityName,
	})

	log.Printf("worker ready — polling %s", taskQueue)
	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalf("worker run: %v", err)
	}
}
