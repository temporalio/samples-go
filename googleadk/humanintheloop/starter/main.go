package main

import (
	"context"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"

	"go.temporal.io/sdk/contrib/googleadk"

	humanintheloop "github.com/temporalio/samples-go/googleadk/humanintheloop"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "google-adk-hitl_workflowID",
		TaskQueue: humanintheloop.TaskQueue,
	}

	request := "Please delete the resource named prod-db."
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, humanintheloop.ApprovalWorkflow, request)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// The agent will call delete_resource, which pauses awaiting human approval.
	// The workflow is now durably blocked on the confirmation signal — it would
	// stay blocked indefinitely (surviving worker restarts) until a decision
	// arrives. Here we simulate a human approving after a short delay.
	log.Println("Waiting for the agent to pause on confirmation, then approving...")
	time.Sleep(3 * time.Second)

	decision := googleadk.ConfirmationDecision{Confirmed: true}
	if err := c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), googleadk.ConfirmationSignalName, decision); err != nil {
		log.Fatalln("Unable to signal approval", err)
	}
	log.Println("Sent approval signal")

	// Synchronously wait for the workflow completion.
	var res humanintheloop.Result
	if err := we.Get(context.Background(), &res); err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}
	log.Printf("Approved=%v answer=%q", res.Approved, res.Answer)
}
