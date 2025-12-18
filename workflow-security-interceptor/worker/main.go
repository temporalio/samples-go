package main

import (
	"log"

	securityinterceptor "github.com/temporalio/samples-go/workflow-security-interceptor"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	sdkinterceptor "go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "security-interceptor", worker.Options{
		Interceptors: []sdkinterceptor.WorkerInterceptor{securityinterceptor.NewWorkerInterceptor()},
	})
	w.RegisterWorkflow(securityinterceptor.Workflow)
	w.RegisterWorkflow(securityinterceptor.ChildWorkflow)
	w.RegisterWorkflow(securityinterceptor.ProhibitedChildWorkflow)
	// Activity used by the interceptor
	w.RegisterActivity(securityinterceptor.ValidateChildWorkflowTypeActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
