package main

import (
	"flag"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"

	"github.com/temporalio/temporal-go-samples/cmd/samples/common"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope:          h.Scope,
		Logger:                h.Logger,
		EnableLoggingInReplay: true,
		EnableSessionWorker:   true,
	}
	workflowWorker := h.StartWorker(h.Config.DomainName, ApplicationName, workerOptions)
	workflowWorker.RegisterWorkflow(SampleFileProcessingWorkflow)

	// Host Specific activities processing case
	workerOptions.DisableWorkflowWorker = true
	worker := h.StartWorker(h.Config.DomainName, HostID, workerOptions)

	worker.RegisterActivityWithOptions(
		downloadFileActivity,
		activity.RegisterOptions{Name: downloadFileActivityName},
	)
	worker.RegisterActivityWithOptions(
		processFileActivity,
		activity.RegisterOptions{Name: processFileActivityName},
	)
	worker.RegisterActivityWithOptions(
		uploadFileActivity,
		activity.RegisterOptions{Name: uploadFileActivityName},
	)
}

func startWorkflow(h *common.SampleHelper, fileID string) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "fileprocessing_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, SampleFileProcessingWorkflow, fileID)
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
		startWorkflow(&h, uuid.New())
	}
}
