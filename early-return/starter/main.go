package main

import (
	"context"
	"log"
	"time"

	"github.com/pborman/uuid"
	"github.com/temporalio/samples-go/early-return"
	enumspb "go.temporal.io/api/enums/v1"
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

	txRequest := earlyreturn.TransactionRequest{SourceAccount: "Bob", TargetAccount: "Alice", Amount: 100}

	startWorkflowOp := c.NewWithStartWorkflowOperation(client.StartWorkflowOptions{
		ID:        "early-return-workflow-ID-" + uuid.New(),
		TaskQueue: earlyreturn.TaskQueueName,
		// WorkflowIDConflictPolicy is required when using UpdateWithStartWorkflow.
		// Here we use FAIL, because we do not expect this workflow ID to exist already,
		// and so we want an error if itdoes.
		WorkflowIDConflictPolicy: enumspb.WORKFLOW_ID_CONFLICT_POLICY_FAIL,
	}, earlyreturn.Workflow, txRequest)

	updateOptions := client.UpdateWorkflowOptions{
		UpdateName:   earlyreturn.UpdateName,
		WaitForStage: client.WorkflowUpdateStageCompleted,
	}

	updateHandle, err := c.UpdateWithStartWorkflow(ctxWithTimeout, client.UpdateWithStartWorkflowOptions{
		StartWorkflowOperation: startWorkflowOp,
		UpdateOptions:          updateOptions,
	})
	if err != nil {
		// For example, a client-side validation error (e.g. missing conflict
		// policy or invalid workflow argument types in the start operation), or
		// a server-side failure (e.g. failed to start workflow, or exceeded
		// limit on concurrent update per workflow execution).
		log.Fatalln("Error issuing update-with-start:", err)
	}

	log.Println("Started workflow",
		"WorkflowID:", updateHandle.WorkflowID(),
		"RunID:", updateHandle.RunID())

	var txInitialization earlyreturn.Transaction
	if err = updateHandle.Get(ctxWithTimeout, &txInitialization); err != nil {
		// The workflow will continue running, cancelling the transaction.

		// NOTE: If the error is retryable, a retry attempt must use a unique workflow ID.
		log.Fatalln("Error obtaining update result:", err)
	}

	workflowRun, err := startWorkflowOp.Get(ctxWithTimeout)
	if err != nil {
		log.Fatalln("Error obtaining workflow run:", err)
	}
	log.Println("Transaction initialized successfully",
		"WorkflowID:", workflowRun.GetID(),
		"RunID:", workflowRun.GetRunID(),
		"TxID:", txInitialization.ID)

	// The workflow will continue running, completing the transaction.

}
