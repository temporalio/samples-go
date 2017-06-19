package main

import (
	"flag"
	"time"

	"github.com/samarabbas/cadence-samples/cmd/samples/common"

	"github.com/pborman/uuid"
	"go.uber.org/cadence"
)

// The cron job can be scheduled with a specified timer interval, if you need cron at a shorter durations less than
// a few minutes on production load, talk to cadence team before doing so. We might have better solutions for you.
var cronSchedule = ScheduleSpec{JobCount: 5, ScheduleInterval: time.Minute * 10}

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := cadence.WorkerOptions{
		MetricsScope: h.Scope,
		Logger:       h.Logger,
	}
	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)
}

//
// To start instance of the workflow.
//
func startWorkflow(h *common.SampleHelper) {
	// This workflow ID can be user business logic identifier as well.
	workflowID := "cron_" + uuid.New()
	workflowOptions := cadence.StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, SampleCronWorkflow, cronSchedule)
}

func main() {
	var mode string
	var intervalInSeconds, jobCount uint
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.UintVar(&intervalInSeconds, "i", 5, "Schedule interval in seconds.")
	flag.UintVar(&jobCount, "c", 3, "Job count to schedule")
	flag.Parse()

	if intervalInSeconds > 0 {
		cronSchedule.ScheduleInterval = time.Second * time.Duration(intervalInSeconds)
	}
	if jobCount > 0 {
		cronSchedule.JobCount = jobCount
	}

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
