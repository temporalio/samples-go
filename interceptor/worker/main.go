package main

import (
	"log"

	"github.com/temporalio/samples-go/interceptor"
	"go.temporal.io/sdk/client"
	sdkinterceptor "go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "interceptor", worker.Options{
		// Create interceptor that will put our tag on the logger
		Interceptors: []sdkinterceptor.WorkerInterceptor{interceptor.NewWorkerInterceptor(interceptor.InterceptorOptions{
			CustomLogTags: map[string]interface{}{"my-custom-key": "my-custom-value"},
		})},
	})

	w.RegisterWorkflow(interceptor.Workflow)
	w.RegisterActivity(interceptor.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
