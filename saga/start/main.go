package main

import (
	"context"
	"log"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/saga"
)

func main() {
	// Create the client object just once per process
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()
	options := client.StartWorkflowOptions{
		ID:        "transfer-money-workflow",
		TaskQueue: saga.TransferMoneyTaskQueue,
	}
	transferDetails := saga.TransferDetails{
		Amount:      54.99,
		FromAccount: "001-001",
		ToAccount:   "002-002",
		ReferenceID: uuid.New().String(),
	}
	we, err := c.ExecuteWorkflow(context.Background(), options, saga.TransferMoney, transferDetails)
	if err != nil {
		log.Fatalln("error starting TransferMoney workflow", err)
	}
	printResults(transferDetails, we.GetID(), we.GetRunID())
}

func printResults(transferDetails saga.TransferDetails, workflowID, runID string) {
	log.Printf(
		"\nTransfer of $%f from account %s to account %s is processing. ReferenceID: %s\n",
		transferDetails.Amount,
		transferDetails.FromAccount,
		transferDetails.ToAccount,
		transferDetails.ReferenceID,
	)
	log.Printf(
		"\nWorkflowID: %s RunID: %s\n",
		workflowID,
		runID,
	)
}
