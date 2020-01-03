package main

import (
	"context"
	"flag"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"

	"github.com/temporalio/temporal-go-samples/cmd/samples/common"
)

const (
	// ApplicationName is the task list for this sample
	ApplicationName = "mutexExample"

	_sampleHelperContextKey = "sampleHelper"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope:              h.Scope,
		Logger:                    h.Logger,
		BackgroundActivityContext: context.WithValue(context.Background(), _sampleHelperContextKey, h),
	}

	// Start Worker.
	worker := worker.New(
		h.Service,
		h.Config.DomainName,
		ApplicationName,
		workerOptions)
	err := worker.Start()
	if err != nil {
		panic("Failed to start workers")
	}
}

// startTwoWorkflows starts two workflows that operate on the same recourceID
func startTwoWorkflows(h *common.SampleHelper) {
	resourceID := uuid.New()
	h.StartWorkflow(client.StartWorkflowOptions{
		ID:                              "SampleWorkflowWithMutex_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    10 * time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	},
		SampleWorkflowWithMutex,
		resourceID)
	h.StartWorkflow(client.StartWorkflowOptions{
		ID:                              "SampleWorkflowWithMutex_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    10 * time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	},
		SampleWorkflowWithMutex,
		resourceID)
}

func main() {
	var mode string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.Parse()

	var h common.SampleHelper
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		startWorkers(&h)

		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "trigger":
		startTwoWorkflows(&h)
	}
}
