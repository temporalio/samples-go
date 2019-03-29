package main

import (
	"flag"
	"github.com/pborman/uuid"
	"github.com/samarabbas/cadence-samples/cmd/samples/common"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"time"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}
	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)
}

func startWorkflow(h *common.SampleHelper) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "query_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Hour * 10,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, QueryWorkflow)
}

func main() {
	var mode, workflowID, runID, queryType string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker, trigger or query.")
	flag.StringVar(&workflowID, "w", "", "WorkflowID")
	flag.StringVar(&runID, "r", "", "RunID")
	flag.StringVar(&queryType, "t", "__stack_trace", "QueryType")
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
	case "query":
		h.QueryWorkflow(workflowID, runID, queryType)
	}
}