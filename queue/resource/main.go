package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"github.com/temporalio/samples-go/queue"
)

func main() {
	// create temporal client
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalln("can't create client", err)
	}
	defer c.Close()

	// create resource pool task queue worker
	resourcePoolWorker := worker.New(c, "resource-pool", worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), queue.ClientContextKey, c),
	})

	// register resource pool related activities and workflows
	resourcePoolWorker.RegisterActivity(queue.SignalWithStartResourcePoolWorkflowActivity)
	resourcePoolWorker.RegisterWorkflow(queue.ResourcePoolWorkflow)
	resourcePoolWorker.RegisterWorkflow(queue.ResourcePoolWorkflowWithInitializer)

	resourcePoolWorker.RegisterActivity(queue.QueryResourcePoolStatusActivity)
	resourcePoolWorker.RegisterActivity(queue.QueryResourceAllocationActivity)

	// start all workers
	workerErr := make(chan error, 1)

	go func() {
		workerErr <- resourcePoolWorker.Run(worker.InterruptCh())
	}()

	// wait for any worker to fail
	err = <-workerErr
	if err != nil {
		log.Fatalln("worker run failed", err)
	}
}
