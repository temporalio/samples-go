package main

import (
	"encoding/gob"
	"flag"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/activity"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/encoded"
	"go.temporal.io/temporal/worker"

	"github.com/temporalio/temporal-go-samples/cmd/samples/common"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope:                       h.Scope,
		Logger:                             h.Logger,
		MaxConcurrentActivityExecutionSize: 1, // Activities are supposed to be CPU intensive, so better limit the concurrency
		DataConverter:                      h.DataConverter,
	}
	worker := h.StartWorker(h.Config.DomainName, ApplicationName, workerOptions)

	worker.RegisterWorkflow(PSOWorkflow)
	worker.RegisterWorkflow(PSOChildWorkflow)

	worker.RegisterActivityWithOptions(
		initParticleActivity,
		activity.RegisterOptions{Name: initParticleActivityName},
	)
	worker.RegisterActivityWithOptions(
		updateParticleActivity,
		activity.RegisterOptions{Name: updateParticleActivityName},
	)
}

func startWorkflow(h *common.SampleHelper, functionName string) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "PSO_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute * 60,
		DecisionTaskStartToCloseTimeout: time.Second * 10, // Measure of responsiveness of the worker to various server signals apart from start workflow. Small means faster recovery in the case of worker failure
	}
	h.StartWorkflow(workflowOptions, PSOWorkflow, functionName)
}

func main() {
	var mode, functionName, workflowID, runID, queryType string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger")
	flag.StringVar(&functionName, "f", "sphere", "One of [sphere, rosenbrock, griewank]")
	flag.StringVar(&workflowID, "w", "", "WorkflowID")
	flag.StringVar(&runID, "r", "", "RunID")
	flag.StringVar(&queryType, "t", "__stack_trace", "Query type is one of [__stack_trace, child, iteration]")
	flag.Parse()

	// If Gob is used to serialize data, then need to register types into gob as well???
	// TOVERIFY: the test works even without type registation!
	const useGob = false
	var dataConverter encoded.DataConverter
	if useGob {
		dataConverter = NewGobDataConverter()
		gob.Register(Vector{})
		gob.Register(Position{})
		gob.Register(Particle{})
		gob.Register(ObjectiveFunction{})
		gob.Register(SwarmSettings{})
		gob.Register(Swarm{})
	} else {
		dataConverter = NewJSONDataConverter()
	}

	var h common.SampleHelper
	h.DataConverter = dataConverter
	h.SetupServiceConfig() // This configures DataConverter

	switch mode {
	case "worker":
		startWorkers(&h)

		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "trigger":
		startWorkflow(&h, functionName)
	case "query":
		h.QueryWorkflow(workflowID, runID, queryType)
	}
}
