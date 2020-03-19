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
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}
	worker := h.StartWorker(h.Config.DomainName, ApplicationName, workerOptions)

	worker.RegisterWorkflow(Workflow)
	worker.RegisterActivity(activityToBeCanceled)
	worker.RegisterActivity(activityToBeSkipped)
	worker.RegisterActivity(cleanupActivity)
}

func startWorkflow(h *common.SampleHelper) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "cancel_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute * 30,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, Workflow)
}

func cancelWorkflow(h *common.SampleHelper, wid string) {
	h.CancelWorkflow(wid)
}

func main() {
	var mode, wid string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker, trigger or cancel.")
	flag.StringVar(&wid, "w", "<workflowID>", "w is the workflowID of the workflow to be canceled.")
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
	case "cancel":
		cancelWorkflow(&h, wid)
	}
}
