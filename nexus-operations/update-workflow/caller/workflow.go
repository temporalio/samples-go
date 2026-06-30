package caller

import (
	"github.com/temporalio/samples-go/nexus-operations/update-workflow/api"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue = "remote-update-caller-tq"
)

// Executes the UpdateWorkflow on the handler namespace's workflow via Nexus Operation
func UpdateRemoteCounterWorkflow(ctx workflow.Context, input api.Input) (api.Output, error) {
	c := workflow.NewNexusClient(api.EndpointName, api.CounterUpdateServiceName)

	fut := c.ExecuteOperation(ctx, api.IncrOperationName, input, workflow.NexusOperationOptions{})

	var exec workflow.NexusOperationExecution // exec token can be used to cancel if the update supports it
	if err := fut.GetNexusOperationExecution().Get(ctx, &exec); err != nil {
		return api.Output{}, err
	}

	var out api.Output
	if err := fut.Get(ctx, &out); err != nil {
		return api.Output{}, err
	}
	return out, nil
}
