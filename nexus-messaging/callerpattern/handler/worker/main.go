package main

import (
	"context"
	"log"
	"os"

	"github.com/nexus-rpc/sdk-go/nexus"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/nexus-messaging/callerpattern/handler"
	"github.com/temporalio/samples-go/nexus-messaging/callerpattern/service"
	"github.com/temporalio/samples-go/nexus/options"
)

const (
	handlerNamespace = "my-target-namespace"
	starterUserID    = "default-user"
)

func main() {
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	clientOptions.Namespace = handlerNamespace

	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	// Start the entity workflow for the default user at boot, using WorkflowIDConflictPolicyUseExisting
	// so it's idempotent.
	workflowID := handler.GetWorkflowID(starterUserID)
	_, err = c.ExecuteWorkflow(context.Background(), client.StartWorkflowOptions{
		ID:                       workflowID,
		TaskQueue:                handler.HandlerTaskQueue,
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING,
	}, handler.GreetingWorkflow, starterUserID)
	if err != nil {
		log.Fatalln("Unable to start entity workflow", err)
	}
	log.Println("Entity workflow started or already running", "WorkflowID", workflowID)

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
