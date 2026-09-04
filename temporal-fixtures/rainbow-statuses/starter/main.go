package main

import (
	"context"
	"time"

	"log"

	"github.com/pborman/uuid"
	rainbowstatuses "github.com/temporalio/samples-go/temporal-fixtures/rainbow-statuses"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

var (
	NumberOfSets = 1

	CustomBoolField     = temporal.NewSearchAttributeKeyBool("CustomBoolField")
	CustomDatetimeField = temporal.NewSearchAttributeKeyTime("CustomDatetimeField")
	CustomDoubleField   = temporal.NewSearchAttributeKeyFloat64("CustomDoubleField")
	CustomIntField      = temporal.NewSearchAttributeKeyInt64("CustomIntField")
	CustomKeywordField  = temporal.NewSearchAttributeKeyKeyword("CustomKeywordField")
	CustomStringField   = temporal.NewSearchAttributeKeyString("CustomStringField")
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{Namespace: "default"})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	i := 1
	for i <= NumberOfSets {
		id := uuid.New()[:6]
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

		for i, s := range statuses {
			workflowOptions := client.StartWorkflowOptions{
				ID:        id + "_" + s.String(),
				TaskQueue: "rainbow-statuses",
				TypedSearchAttributes: temporal.NewSearchAttributes(
					CustomKeywordField.ValueSet("rainbow-statuses-"+id),
					CustomIntField.ValueSet(int64(i)),
					CustomDoubleField.ValueSet(float64(i)),
					CustomBoolField.ValueSet(i%2 == 0),
					CustomDatetimeField.ValueSet(time.Now().UTC()),
					CustomStringField.ValueSet("rainbow statuses "+id+" "+s.String()),
				),
			}

			if s == enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT {
				workflowOptions.WorkflowExecutionTimeout = time.Second
			}

			w, err := c.ExecuteWorkflow(context.Background(), workflowOptions,
				rainbowstatuses.RainbowStatusesWorkflow, s)

			if err != nil {
				log.Fatalln("Unable to execute workflow", err)
			}

			switch s {
			case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
				if err = c.CancelWorkflow(context.Background(), w.GetID(), w.GetRunID()); err != nil {
					log.Fatalln("Unable to Cancel workflow", err)
				}
			case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
				if err = c.TerminateWorkflow(context.Background(), w.GetID(), w.GetRunID(), "sample termination"); err != nil {
					log.Fatalln("Unable to Terminate workflow", err)
				}
			}

			if s == enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
				signal := struct {
					Hey string
					At  time.Time
				}{"from Mars", time.Now()}
				if err = c.SignalWorkflow(context.Background(), w.GetID(), w.GetRunID(), "customSignal", signal); err != nil {
					log.Fatalln("unable to signal workflow", err)
				}
			}
		}
	}
}
