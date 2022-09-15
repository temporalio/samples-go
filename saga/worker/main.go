package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/saga"
)

func main() {
	// Create the client object just once per process
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()
	// This worker hosts both Workflow and Activity functions
	w := worker.New(c, saga.TransferMoneyTaskQueue, worker.Options{})
	w.RegisterWorkflow(saga.TransferMoney)
	w.RegisterActivity(saga.Withdraw)
	w.RegisterActivity(saga.WithdrawCompensation)
	w.RegisterActivity(saga.Deposit)
	w.RegisterActivity(saga.DepositCompensation)
	w.RegisterActivity(saga.StepWithError)
	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}
