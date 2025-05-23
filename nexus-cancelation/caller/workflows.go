package caller

import (
	"errors"

	"github.com/temporalio/samples-go/nexus/service"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	TaskQueue    = "my-caller-workflow-task-queue"
	endpointName = "my-nexus-endpoint-name"
)

var languages = []service.Language{
	service.EN,
	service.FR,
	service.DE,
	service.ES,
	service.TR,
}

type result struct {
	lang   service.Language
	output service.HelloOutput
	err    error
}

func HelloCallerWorkflow(ctx workflow.Context, name string) (string, error) {
	log := workflow.GetLogger(ctx)
	c := workflow.NewNexusClient(endpointName, service.HelloServiceName)
	results := make([]result, len(languages))
	callCtx, cancel := workflow.WithCancel(ctx)
	defer cancel()

	// Concurrently execute an operation per language.
	wg := workflow.NewWaitGroup(ctx)
	for i, lang := range languages {
		wg.Add(1)
		// Use the cancelable callCtx for executing the operation.
		workflow.Go(callCtx, func(ctx workflow.Context) {
			defer wg.Done()
			fut := c.ExecuteOperation(
				ctx,
				service.HelloOperationName,
				service.HelloInput{Name: name, Language: lang},
				workflow.NexusOperationOptions{})
			var output service.HelloOutput
			// This future gets resolved when the operation completes with either success, failure, timeout,
			// or cancelation.
			// Only asynchronous operations may receive cancelation as cancelation in Nexus is sent using an
			// operation token.
			// The workflow or any other underlying resource backing the operation may choose to ignore the
			// cancelation request, allowing the operation to end up in any of the terminal states.
			err := fut.Get(ctx, &output)
			results[i] = result{
				lang:   lang,
				output: output,
				err:    err,
			}
			// The first operation to win the race cancels the others.
			cancel()
		})
	}
	// Wait for all operations to resolve. Once the workflow completes, the server will stop trying to cancel any
	// operations that have not yet received cancelation, letting them run to completion. It is totally valid to
	// abandon operations for certain use cases.
	wg.Wait(ctx)

	var greeting string
	for _, res := range results {
		if res.err != nil {
			// Only return an error if an operation errored out not due to cancelation.
			if errors.As(res.err, new(*temporal.CanceledError)) {
				log.Info("operation canceled", "operation", service.HelloOperationName, "lang", res.lang)
				continue
			}
			return "", res.err
		}
		greeting = res.output.Message
	}
	return greeting, nil
}
