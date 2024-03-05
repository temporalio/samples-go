package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	async_update "github.com/temporalio/samples-go/async-update"
	enumspb "go.temporal.io/api/enums/v1"
	updatepb "go.temporal.io/api/update/v1"

	"github.com/temporalio/samples-go/update"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "async-update-workflow-ID",
		TaskQueue: "async-update",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, async_update.ProcessWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Send multiple updates to the workflow
	var updates []client.WorkflowUpdateHandle
	// ProcessWorkflow only allows 5 in progress jobs at a time, so we send 6 updates to test the validator.
	for i := 0; i < 6; i++ {
		updateID := uuid.New().String()
		log.Println("Sending workflow update", "WorkflowID", we.GetID(), "RunID", we.GetRunID(), "UpdateID", updateID)
		handle, err := c.UpdateWorkflowWithOptions(context.Background(), &client.UpdateWorkflowWithOptionsRequest{
			WorkflowID: we.GetID(),
			RunID:      we.GetRunID(),
			UpdateID:   updateID,
			UpdateName: async_update.ProcessUpdateName,
			Args:       []interface{}{"world"},
			// WaitPolicy is a hint to return early if the update reaches a certain stage.
			// By default the SDK will wait until the update is processed or the server sends back
			// an empty response then the SDK can poll the update result later.
			// Useful for short updates that can be completed with a single RPC.
			WaitPolicy: &updatepb.WaitPolicy{
				// LifecycleStage UPDATE_WORKFLOW_EXECUTION_LIFECYCLE_STAGE_ACCEPTED means to wait until the update is accepted
				// or the Temporal server returns an empty response.
				LifecycleStage: enumspb.UPDATE_WORKFLOW_EXECUTION_LIFECYCLE_STAGE_ACCEPTED,
			},
		})
		if err != nil {
			log.Fatalln("Unable to send update request", err)
		}
		updates = append(updates, handle)
	}
	for _, handle := range updates {
		var updateOutput string
		err = handle.Get(context.Background(), &updateOutput)
		if err != nil {
			log.Println("Update failed with error", err)
		} else {
			log.Println("Update result", "WorkflowID", we.GetID(), "RunID", we.GetRunID(), "UpdateID", handle.UpdateID(), "Result", updateOutput)
		}
	}
	// You can also create a handle for a previously sent update using the update's ID.
	newHandle := c.GetWorkflowUpdateHandle(client.GetWorkflowUpdateHandleOptions{
		WorkflowID: we.GetID(),
		RunID:      we.GetRunID(),
		UpdateID:   updates[0].UpdateID(),
	})
	var updateOutput string
	err = newHandle.Get(context.Background(), &updateOutput)
	if err != nil {
		log.Println("Get update result failed with error", err)
	} else {
		log.Println("Get update result", "WorkflowID", we.GetID(), "RunID", we.GetRunID(), "UpdateID", newHandle.UpdateID(), "Result", updateOutput)
	}
	// Signal the workflow to stop accepting new work.
	if err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), update.Done, nil); err != nil {
		log.Fatalf("failed to send %q signal to workflow: %v", update.Done, err)
	}
	// Get the result of the workflow this will block until all the updates are processed.
	var wfResult int
	if err = we.Get(context.Background(), &wfResult); err != nil {
		log.Fatalf("unable get workflow result: %v", err)
	}
	log.Println("Updates processed:", wfResult)
}
