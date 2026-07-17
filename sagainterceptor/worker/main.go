package main

import (
	"log"

	"github.com/temporalio/samples-go/saga"
	"github.com/temporalio/samples-go/sagainterceptor"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	// Create the client object just once per process
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	intercept, _ := sagainterceptor.NewInterceptor(sagainterceptor.InterceptorOptions{
		WorkflowRegistry: map[string]sagainterceptor.SagaOptions{
			"TransferMoney": {},
		},
		ActivityRegistry: map[string]sagainterceptor.CompensationOptions{
			"Withdraw": {
				ActivityType: "WithdrawCompensation",
			},
			"Deposit": {
				ActivityType: "DepositCompensation",
			},
		},
	})
	// This worker hosts both Workflow and Activity functions
	w := worker.New(c, sagainterceptor.TransferMoneyTaskQueue, worker.Options{
		Interceptors: []interceptor.WorkerInterceptor{intercept},
	})
	w.RegisterWorkflowWithOptions(sagainterceptor.TransferMoney, workflow.RegisterOptions{
		Name: "TransferMoney",
	})
	w.RegisterActivityWithOptions(sagainterceptor.Withdraw, activity.RegisterOptions{
		Name: "Withdraw",
	})
	w.RegisterActivityWithOptions(sagainterceptor.WithdrawCompensation, activity.RegisterOptions{
		Name: "WithdrawCompensation",
	})
	w.RegisterActivityWithOptions(sagainterceptor.Deposit, activity.RegisterOptions{
		Name: "Deposit",
	})
	w.RegisterActivityWithOptions(sagainterceptor.DepositCompensation, activity.RegisterOptions{
		Name: "DepositCompensation",
	})
	w.RegisterActivity(saga.StepWithError)
	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}
