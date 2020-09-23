package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/recovery"
)

func main() {
	var workflowID, input, workflowType string
	flag.StringVar(&workflowID, "w", "trip_workflow", "WorkflowID.")
	flag.StringVar(&input, "i", "{}", "Workflow input parameters.")
	flag.StringVar(&workflowType, "wt", "tripworkflow", "Workflow type (tripworkflow|recoveryworkflow).")
	flag.Parse()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	var we client.WorkflowRun
	var weError error
	switch workflowType {
	case "tripworkflow":
		var userState recovery.UserState
		if err := json.Unmarshal([]byte(input), &userState); err != nil {
			log.Fatalln("Unable to unmarshal workflow input parameters", err)
		}
		workflowOptions := client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: "recovery",
		}
		we, weError = c.ExecuteWorkflow(context.Background(), workflowOptions, recovery.TripWorkflow, userState)
	case "recoveryworkflow":
		var params recovery.Params
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			log.Fatalln("Unable to unmarshal workflow input parameters", err)
		}

		workflowOptions := client.StartWorkflowOptions{
			ID:        workflowID,
			TaskQueue: "recovery",
		}
		we, weError = c.ExecuteWorkflow(context.Background(), workflowOptions, recovery.RecoverWorkflow, params)
	default:
		flag.PrintDefaults()
		return
	}

	if weError != nil {
		log.Fatalln("Unable to execute workflow", err)
	} else {
		log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
	}
}
