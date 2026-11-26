package main

import (
	"context"
	"log"
	"time"

	update_cancel "github.com/temporalio/samples-go/update-cancel"
	enumspb "go.temporal.io/api/enums/v1"
	updatepb "go.temporal.io/api/update/v1"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "update_cancel-workflow-ID",
		TaskQueue: "update_cancel",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, update_cancel.UpdateWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	cancellableUpdateID := "cancellable-update-ID"
	log.Println("Sending update", "UpdateID", cancellableUpdateID)

	// Send an async update request.
	handle, err := c.UpdateWorkflowWithOptions(context.Background(), &client.UpdateWorkflowWithOptionsRequest{
		WorkflowID: we.GetID(),
		RunID:      we.GetRunID(),
		UpdateName: update_cancel.UpdateHandle,
		UpdateID:   cancellableUpdateID,
		WaitPolicy: &updatepb.WaitPolicy{
			LifecycleStage: enumspb.UPDATE_WORKFLOW_EXECUTION_LIFECYCLE_STAGE_ACCEPTED,
		},
		Args: []interface{}{
			4 * time.Hour,
		},
	})
	if err != nil {
		log.Fatalln("Unable to execute update", err)
	}
	log.Println("Sent update")

	log.Println("Waiting 5s to send cancel")
	time.Sleep(5 * time.Second)
	log.Println("Sending cancel to update", "UpdateID", cancellableUpdateID)

	_, err = c.UpdateWorkflow(context.Background(), we.GetID(), we.GetRunID(), update_cancel.UpdateCancelHandle, cancellableUpdateID)
	if err != nil {
		log.Fatalln("Unable to send cancel", err)
	}
	log.Println("Sent cancel")

	var sleepTime time.Duration
	err = handle.Get(context.Background(), &sleepTime)
	if err != nil {
		log.Fatalln("Unable to get update result", err)
	}
	// Update will only sleep for 5s because it was cancelled.
	log.Println("Update slept for:", sleepTime)

	if err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), update_cancel.Done, nil); err != nil {
		log.Fatalf("failed to send %q signal to workflow: %v", update_cancel.Done, err)
	}
	var wfresult int
	if err = we.Get(context.Background(), &wfresult); err != nil {
		log.Fatalf("unable get workflow result: %v", err)
	}
	log.Println("workflow result:", wfresult)
}
