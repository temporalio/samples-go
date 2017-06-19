package main

import (
	"flag"
	"time"

	"github.com/samarabbas/cadence-samples/cmd/samples/common"

	"github.com/pborman/uuid"
	"go.uber.org/cadence"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := cadence.WorkerOptions{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}

	// Start Worker.
	worker := cadence.NewWorker(
		h.Service,
		h.Config.DomainName,
		ApplicationName,
		workerOptions)
	err := worker.Start()
	if err != nil {
		panic("Failed to start workers")
	}
}

func startWorkflow(h *common.SampleHelper) {
	workflowOptions := cadence.StartWorkflowOptions{
		ID:                              "pickfirst_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, SamplePickFirstWorkflow)
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
		startWorkflow(&h)
	}
}
