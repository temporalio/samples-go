package main

import (
	"context"
	"log"
	"time"

	"github.com/temporalio/samples-go/interceptor"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	sdkinterceptor "go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "interceptor", worker.Options{
		// Create interceptor that will put started time on the logger
		Interceptors: []sdkinterceptor.WorkerInterceptor{interceptor.NewWorkerInterceptor(interceptor.InterceptorOptions{
			GetExtraLogTagsForWorkflow: func(ctx workflow.Context) []interface{} {
				return []interface{}{"WorkflowStartTime", workflow.GetInfo(ctx).WorkflowStartTime.Format(time.RFC3339)}
			},
			GetExtraLogTagsForActivity: func(ctx context.Context) []interface{} {
				return []interface{}{"ActivityStartTime", activity.GetInfo(ctx).StartedTime.Format(time.RFC3339)}
			},
		})},
	})

	w.RegisterWorkflow(interceptor.Workflow)
	w.RegisterActivity(interceptor.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
