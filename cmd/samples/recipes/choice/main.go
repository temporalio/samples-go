package main

import (
	"flag"
	"time"

	"github.com/pborman/uuid"
	"github.com/samarabbas/cadence-samples/cmd/samples/common"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}

	// Start Worker.
	worker := worker.NewWorker(
		h.Service,
		h.Config.DomainName,
		ApplicationName,
		workerOptions)
	err := worker.Start()
	if err != nil {
		panic("Failed to start workers")
	}
}

func startWorkflowMultiChoice(h *common.SampleHelper) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "multi_choice_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, MultiChoiceWorkflow)
}

func startWorkflowExclusiveChoice(h *common.SampleHelper) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "single_choice_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, ExclusiveChoiceWorkflow)
}

func main() {
	var mode, sampleCase string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.StringVar(&sampleCase, "c", "single", "Sample case to run.")
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
		switch sampleCase {
		case "multi":
			startWorkflowMultiChoice(&h)
		default:
			startWorkflowExclusiveChoice(&h)
		}
	}
}
