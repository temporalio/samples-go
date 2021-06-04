package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	uidriven "github.com/temporalio/samples-go/ui-driven"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "ui-driven", worker.Options{})

	w.RegisterWorkflow(uidriven.OrderWorkflow)
	w.RegisterWorkflow(uidriven.UpdateOrderWorkflow)
	w.RegisterWorkflow(uidriven.ShippingWorkflow)
	w.RegisterActivity(uidriven.RegisterEmail)
	w.RegisterActivity(uidriven.ValidateSize)
	w.RegisterActivity(uidriven.ValidateColor)
	w.RegisterActivity(uidriven.ScheduleDelivery)
	w.RegisterActivity(uidriven.SendDeliveryEmail)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
