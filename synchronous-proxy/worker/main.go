package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	synchronousproxy "github.com/temporalio/samples-go/synchronous-proxy"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "ui-driven", worker.Options{})

	w.RegisterWorkflow(synchronousproxy.OrderWorkflow)
	w.RegisterWorkflow(synchronousproxy.UpdateOrderWorkflow)
	w.RegisterWorkflow(synchronousproxy.ShippingWorkflow)
	w.RegisterActivity(synchronousproxy.RegisterEmail)
	w.RegisterActivity(synchronousproxy.ValidateSize)
	w.RegisterActivity(synchronousproxy.ValidateColor)
	w.RegisterActivity(synchronousproxy.ScheduleDelivery)
	w.RegisterActivity(synchronousproxy.SendDeliveryEmail)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
