package main

import (
	"context"
	"log"
	"time"

	"github.com/temporalio/samples-go/logger-interceptor"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	sdkinterceptor "go.temporal.io/sdk/interceptor"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "logger-interceptor", worker.Options{
		// Create logger-interceptor that will put started time on the logger
		Interceptors: []sdkinterceptor.WorkerInterceptor{logger_interceptor.NewWorkerInterceptor(logger_interceptor.InterceptorOptions{
			GetExtraLogTagsForWorkflow: func(ctx workflow.Context) []interface{} {
				return []interface{}{"WorkflowStartTime", workflow.GetInfo(ctx).WorkflowStartTime.Format(time.RFC3339)}
			},
			GetExtraLogTagsForActivity: func(ctx context.Context) []interface{} {
				return []interface{}{"ActivityStartTime", activity.GetInfo(ctx).StartedTime.Format(time.RFC3339)}
			},
		})},
	})

	w.RegisterWorkflow(logger_interceptor.Workflow)
	w.RegisterActivity(logger_interceptor.Activity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
