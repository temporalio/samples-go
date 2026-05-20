package main

import (
	"context"
	"log"

	externalstorage "github.com/temporalio/samples-go/external-storage"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := externalstorage.NewClient(context.Background(), client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, externalstorage.TaskQueue, worker.Options{})
	w.RegisterWorkflow(externalstorage.ProcessOrderBatchWorkflow)
	w.RegisterActivity(externalstorage.FetchOrders)
	w.RegisterActivity(externalstorage.ProcessOrders)

	if err := w.Run(worker.InterruptCh()); err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
