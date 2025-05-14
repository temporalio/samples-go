// @@@SNIPSTART samples-go-nexus-caller-workflow
package caller

import (
	"time"

	"github.com/temporalio/samples-go/nexus/service"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue    = "my-caller-workflow-task-queue"
	endpointName = "my-nexus-endpoint-name"
)

func EchoCallerWorkflow(ctx workflow.Context, message string) (string, error) {
	c := workflow.NewNexusClient(endpointName, service.HelloServiceName)

	fut := c.ExecuteOperation(ctx, service.EchoOperationName, service.EchoInput{Message: message}, workflow.NexusOperationOptions{})

	var res service.EchoOutput
	if err := fut.Get(ctx, &res); err != nil {
		return "", err
	}

	return res.Message, nil
}

func HelloCallerWorkflow(ctx workflow.Context, name string, language service.Language) (string, error) {
	c := workflow.NewNexusClient(endpointName, service.HelloServiceName)

	fut := c.ExecuteOperation(ctx, service.HelloOperationName, service.HelloInput{Name: name, Language: language}, workflow.NexusOperationOptions{})

	// Optionally wait for the operation to be started. NexusOperationExecution will contain the operation token in
	// case this operation is asynchronous, which is a handle that can be used to perform additional actions like
	// cancelling an operation.
	var exec workflow.NexusOperationExecution
	if err := fut.GetNexusOperationExecution().Get(ctx, &exec); err != nil {
		return "", err
	}

	var res service.HelloOutput
	if err := fut.Get(ctx, &res); err != nil {
		return "", err
	}

	return res.Message, nil
}

func CancelCallerWorkflow(ctx workflow.Context) error {
	cctx, cancel := workflow.WithCancel(ctx)
	c := workflow.NewNexusClient(endpointName, service.HelloServiceName)
	fut := c.ExecuteOperation(cctx, service.CancelOperationName, nil, workflow.NexusOperationOptions{})
	var exec workflow.NexusOperationExecution
	if err := fut.GetNexusOperationExecution().Get(ctx, &exec); err != nil {
		return err
	}

	if err := workflow.Sleep(ctx, 60*time.Second); err != nil {
		return err
	}

	cancel()
	if err := fut.Get(ctx, nil); err != nil {
		return err
	}

	return nil
}

// @@@SNIPEND
