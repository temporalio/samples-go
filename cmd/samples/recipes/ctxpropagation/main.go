package main

import (
	"context"
	"flag"
	"time"

	"github.com/pborman/uuid"
	"github.com/samarabbas/cadence-samples/cmd/samples/common"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options. Setup two context propagators - one string and
	// one custom context.
	workerOptions := worker.Options{
		MetricsScope:          h.Scope,
		Logger:                h.Logger,
		EnableLoggingInReplay: true,
		ContextPropagators: []workflow.ContextPropagator{
			NewContextPropagator(),
		},
	}
	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)

	// Host Specific activities processing case
	workerOptions.DisableWorkflowWorker = true
	h.StartWorkers(h.Config.DomainName, HostID, workerOptions)
}

func startWorkflow(h *common.SampleHelper) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "ctxprop_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, propagateKey, &Values{Key: "test", Value: "tested"})
	h.StartWorkflowWithCtx(ctx, workflowOptions, CtxPropWorkflow)
}

func main() {
	var mode string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.Parse()

	var h common.SampleHelper
	// Setup two context propagators - one string and one custom context.
	h.CtxPropagators = []workflow.ContextPropagator{
		NewContextPropagator(),
	}
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		startWorkers(&h)

		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "trigger":
		startWorkflow(&h)
	}
}
