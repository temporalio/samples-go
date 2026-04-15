package main

import (
	"context"
	"log"

	"github.com/nexus-rpc/sdk-go/nexus"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus-messaging/callerpattern/handler"
	"github.com/temporalio/samples-go/nexus-messaging/callerpattern/service"
)

const starterUserID = "default-user"

func main() {
	// Connect to the handler's target namespace. For a non-local setup, provide additional
	// client options such as HostPort and TLS credentials.
	c, err := client.Dial(client.Options{Namespace: "my-target-namespace"})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// Start the GreetingWorkflow for the default user at boot, using WorkflowIDConflictPolicyUseExisting
	// so it's idempotent.
	workflowID := handler.GetWorkflowID(starterUserID)
	_, err = c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                handler.HandlerTaskQueue,
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
	}, handler.GreetingWorkflow, starterUserID)
	if err != nil {
		log.Fatalln("Unable to start caller workflow", err)
	}
	log.Println("GreetingWorkflow started or already running", "WorkflowID", workflowID)

	w := worker.New(c, handler.HandlerTaskQueue, worker.Options{})

	svc := nexus.NewService(service.ServiceName)
	err = svc.Register(
		handler.GetLanguagesOperation,
		handler.GetLanguageOperation,
		handler.SetLanguageOperation,
		handler.ApproveOperation,
	)
	if err != nil {
		log.Fatalln("Unable to register operations", err)
	}
	w.RegisterNexusService(svc)
	w.RegisterWorkflow(handler.GreetingWorkflow)
	w.RegisterActivity(handler.GreetingActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
