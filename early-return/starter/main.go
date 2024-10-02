package main

import (
	"context"
	"log"
	"time"

	"github.com/pborman/uuid"
	"github.com/temporalio/samples-go/early-return"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateOperation := client.NewUpdateWithStartWorkflowOperation(
		client.UpdateWorkflowOptions{
			UpdateName:   earlyreturn.UpdateName,
			WaitForStage: client.WorkflowUpdateStageCompleted,
		})

	txId := uuid.New()
	workflowOptions := client.StartWorkflowOptions{
		ID:                 "early-return-workflow-ID-" + txId,
		TaskQueue:          earlyreturn.TaskQueueName,
		WithStartOperation: updateOperation,
	}
	we, err := c.ExecuteWorkflow(ctxWithTimeout, workflowOptions, earlyreturn.Workflow, txId, "bob", "alice", 100.0)
	if err != nil {
		log.Fatalln("Error executing workflow:", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	updateHandle, err := updateOperation.Get(ctxWithTimeout)
	if err != nil {
		log.Fatalln("Error obtaining update handle:", err)
	}

	err = updateHandle.Get(ctxWithTimeout, nil)
	if err != nil {
		// NOTE: If the error is retryable, a retry attempt must use a unique workflow ID.
		log.Fatalln("Error obtaining update result:", err)
	}

	log.Println("Transaction completed successfully")

	// The workflow will continue running, either completing or cancelling the transaction.
}
