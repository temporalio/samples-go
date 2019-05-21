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

var (
	ctxVals = map[string]string{
		"key1": "abc",
		"key2": "def",
		"key3": "ghi",
	}
	propagatedVals = map[string]string{
		"key1": "abc",
		"key2": "def",
	}
	propagatedKeys = []string{"key1", "key2"}
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
			workflow.NewStringMapPropagator(propagatedKeys),
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
	for key, val := range ctxVals {
		ctx = context.WithValue(ctx, workflow.ContextKey(key), val)
	}
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
		workflow.NewStringMapPropagator(propagatedKeys),
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
