package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/temporalio/temporal-go-samples/cmd/samples/common"
	"github.com/temporalio/temporal-go-samples/cmd/samples/recovery/cache"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"go.uber.org/zap"
	"time"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}
	ctx := context.WithValue(context.Background(), TemporalClientKey, workflowClient)
	ctx = context.WithValue(ctx, WorkflowExecutionCacheKey, cache.NewLRU(10))

	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
		BackgroundActivityContext: ctx,
	}

	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)

	// Configure worker options.
	hostSpecificWorkerOptions := worker.Options{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
		BackgroundActivityContext: ctx,
		DisableWorkflowWorker: true,
	}

	h.StartWorkers(h.Config.DomainName, HostID, hostSpecificWorkerOptions)
}

func startTripWorkflow(h *common.SampleHelper, workflowID string, user UserState) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Hour * 24,
		DecisionTaskStartToCloseTimeout: time.Second * 10,
	}
	h.StartWorkflow(workflowOptions, TripWorkflow, user)
}

func startRecoveryWorkflow(h *common.SampleHelper, workflowID string, params Params) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Hour * 24,
		DecisionTaskStartToCloseTimeout: time.Second * 10,
	}
	h.StartWorkflow(workflowOptions, RecoverWorkflow, params)
}

func main() {
	var mode, workflowID,signal, input, workflowType string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.StringVar(&workflowID, "w", "workflow_A", "WorkflowID")
	flag.StringVar(&signal, "s", "signal_data", "SignalData")
	flag.StringVar(&input, "i", "{}", "Workflow input parameters.")
	flag.StringVar(&workflowType, "wt", "main.TripWorkflow", "Workflow type.")
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
		switch workflowType {
		case "tripworkflow":
			var userState UserState
			if err := json.Unmarshal([]byte(input), &userState); err != nil {
				panic(err)
			}
			startTripWorkflow(&h, workflowID, userState)
		case "recoveryworkflow":
			var params Params
			if err := json.Unmarshal([]byte(input), &params); err != nil {
				panic(err)
			}
			startRecoveryWorkflow(&h, workflowID, params)
		}
	case "query":
		h.QueryWorkflow(workflowID, "", QueryName)
	case "signal":
		var tripEvent TripEvent
		if err := json.Unmarshal([]byte(signal), &tripEvent); err != nil {
			panic(err)
		}
		h.SignalWorkflow(workflowID, TripSignalName, tripEvent)
	}
}
