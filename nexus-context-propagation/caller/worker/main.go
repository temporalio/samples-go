package main

import (
	"log"
	"os"

	"github.com/temporalio/samples-go/ctxpropagation"
	nexuscontextpropagation "github.com/temporalio/samples-go/nexus-context-propagation"
	"github.com/temporalio/samples-go/nexus/caller"
	"github.com/temporalio/samples-go/nexus/options"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	clientOptions, err := options.ParseClientOptionFlags(os.Args[1:])
	if err != nil {
		log.Fatalf("Invalid arguments: %v", err)
	}
	// Set up context propagators from workflow and non-workflow contexts.
	clientOptions.ContextPropagators = []workflow.ContextPropagator{ctxpropagation.NewContextPropagator()}
	c, err := client.Dial(clientOptions)
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, caller.TaskQueue, worker.Options{
		Interceptors: []interceptor.WorkerInterceptor{
			&nexuscontextpropagation.WorkerInterceptor{
				// Use the provided data converter to encode the Nexus headers. Use a custom data
				// converter to encrypt the header values.
				// IMPORTANT: Nexus headers values are plain strings and are not visited by the
				// grpc-proxy (see related sample), special care should be taken when used to pass
				// sensitive information.
				DataConverter: converter.GetDefaultDataConverter(),
			},
		},
	})

	w.RegisterWorkflow(caller.HelloCallerWorkflow)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
