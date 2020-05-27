package main

import (
	"context"
	"encoding/json"
	"flag"

	"go.temporal.io/temporal/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/recovery"
)

func main() {
	var workflowID, input, workflowType string
	flag.StringVar(&workflowID, "w", "trip_workflow", "WorkflowID.")
	flag.StringVar(&input, "i", "{}", "Workflow input parameters.")
	flag.StringVar(&workflowType, "wt", "tripworkflow", "Workflow type (tripworkflow|recoveryworkflow).")
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer c.Close()

	var we client.WorkflowRun
	var weError error
	switch workflowType {
	case "tripworkflow":
		var userState recovery.UserState
		if err := json.Unmarshal([]byte(input), &userState); err != nil {
			logger.Fatal("Unable to unmarshal workflow input parameters", zap.Error(err))
		}
		workflowOptions := client.StartWorkflowOptions{
			ID:       workflowID,
			TaskList: "recovery",
		}
		we, weError = c.ExecuteWorkflow(context.Background(), workflowOptions, recovery.TripWorkflow, userState)
	case "recoveryworkflow":
		var params recovery.Params
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			logger.Fatal("Unable to unmarshal workflow input parameters", zap.Error(err))
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:       workflowID,
			TaskList: "recovery",
		}
		we, weError = c.ExecuteWorkflow(context.Background(), workflowOptions, recovery.RecoverWorkflow, params)
	default:
		flag.PrintDefaults()
		return
	}

	if weError != nil {
		logger.Error("Unable to execute workflow", zap.Error(err))
	} else {
		logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
	}
}
