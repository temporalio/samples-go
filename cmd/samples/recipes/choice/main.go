package main

import (
	"flag"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"

	"github.com/temporalio/temporal-go-samples/cmd/samples/common"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		HostPort:     h.Config.HostPort,
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}

	// Start Worker.
	worker, err := worker.New(
		h.Config.DomainName,
		ApplicationName,
		workerOptions)
	if err != nil {
		panic("Failed to create workers")
	}
	err = worker.Start()
	if err != nil {
		panic("Failed to start workers")
	}

	worker.RegisterWorkflow(ExclusiveChoiceWorkflow)
	worker.RegisterActivity(getOrderActivity)
	worker.RegisterActivity(orderAppleActivity)
	worker.RegisterActivity(orderBananaActivity)
	worker.RegisterActivity(orderCherryActivity)
	worker.RegisterActivity(orderOrangeActivity)
	worker.RegisterWorkflow(MultiChoiceWorkflow)
	worker.RegisterActivity(getBasketOrderActivity)
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
