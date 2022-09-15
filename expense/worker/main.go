package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/expense"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "expense", worker.Options{})

	w.RegisterWorkflow(expense.SampleExpenseWorkflow)
	w.RegisterActivity(expense.CreateExpenseActivity)
	w.RegisterActivity(expense.WaitForDecisionActivity)
	w.RegisterActivity(expense.PaymentActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
