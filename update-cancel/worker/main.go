package main

import (
	"log"

	update_cancel "github.com/temporalio/samples-go/update-cancel"
	"go.temporal.io/sdk/client"
	sdkinterceptor "go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "update_cancel", worker.Options{
		Interceptors: []sdkinterceptor.WorkerInterceptor{update_cancel.NewWorkerInterceptor()},
	})

	w.RegisterWorkflow(update_cancel.UpdateWorkflow)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
