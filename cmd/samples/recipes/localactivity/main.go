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

	worker.RegisterWorkflow(ProcessingWorkflow)
	worker.RegisterWorkflow(SignalHandlingWorkflow)
	worker.RegisterActivity(activityForCondition0)
	worker.RegisterActivity(activityForCondition1)
	worker.RegisterActivity(activityForCondition2)
	worker.RegisterActivity(activityForCondition3)
	worker.RegisterActivity(activityForCondition4)

	// no need to register local activities
}

func startWorkflow(h *common.SampleHelper) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "localactivity_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute * 3,
		DecisionTaskStartToCloseTimeout: time.Minute,
		WorkflowIDReusePolicy:           client.WorkflowIDReusePolicyAllowDuplicate,
	}
	h.StartWorkflow(workflowOptions, SignalHandlingWorkflow)
}

func main() {
	var mode, workflowID, signal string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker, trigger or query.")
	flag.StringVar(&workflowID, "w", "", "WorkflowID")
	flag.StringVar(&signal, "s", "signal_data", "SignalData")
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
	case "signal":
		h.SignalWorkflow(workflowID, SignalName, signal)
	}
}
