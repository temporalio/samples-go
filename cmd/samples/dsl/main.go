package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/worker"
	"gopkg.in/yaml.v2"

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
	worker.RegisterWorkflow(SimpleDSLWorkflow)
	worker.RegisterActivity(sampleActivity1)
	worker.RegisterActivity(sampleActivity2)
	worker.RegisterActivity(sampleActivity3)
	worker.RegisterActivity(sampleActivity4)
	worker.RegisterActivity(sampleActivity5)
}

func startWorkflow(h *common.SampleHelper, w Workflow) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "dsl_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, SimpleDSLWorkflow, w)
}

func main() {
	var mode, dslConfig string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.StringVar(&dslConfig, "dslConfig", "cmd/samples/dsl/workflow1.yaml", "dslConfig specify the yaml file for the dsl workflow.")
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

		data, err := ioutil.ReadFile(dslConfig)
		if err != nil {
			panic(fmt.Sprintf("failed to load dsl config file %v", err))
		}
		var workflow Workflow
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			panic(fmt.Sprintf("failed to unmarshal dsl config %v", err))
		}
		startWorkflow(&h, workflow)
	}
}
