package main

import (
	"context"
	"time"

	// "fmt"
	"log"

	"github.com/pborman/uuid"
	rainbowstatuses "github.com/temporalio/samples-go/temporal-fixtures/rainbow-statuses"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

var (
	NumberOfSets = 1
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{Namespace: "default"})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	uuidvar := uuid.New()
	i := 1
	for i <= NumberOfSets {
		id := uuidvar[:6]
		i++

		statuses := []enums.WorkflowExecutionStatus{
			enums.WORKFLOW_EXECUTION_STATUS_RUNNING,
			enums.WORKFLOW_EXECUTION_STATUS_COMPLETED,
			enums.WORKFLOW_EXECUTION_STATUS_FAILED,
			enums.WORKFLOW_EXECUTION_STATUS_CANCELED,
			enums.WORKFLOW_EXECUTION_STATUS_TERMINATED,
			enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW,
			enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
		}

		for _, s := range statuses {

			workflowOptions := client.StartWorkflowOptions{
				ID:        id + "_" + s.String(),
				TaskQueue: "rainbow-statuses",
			}

			if s == enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT {
				workflowOptions.WorkflowExecutionTimeout = 1 * time.Second
			}

			w, err := c.ExecuteWorkflow(context.Background(), workflowOptions,
				rainbowstatuses.RainbowStatusesWorkflow, s)

			if err != nil {
				log.Fatalln("Unable to execute workflow", err)
			}

			switch s {
			case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
				c.CancelWorkflow(context.Background(), w.GetID(), w.GetRunID())
			case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
				c.TerminateWorkflow(context.Background(), w.GetID(), w.GetRunID(), "sample termination")

			}

		}

	}
}
