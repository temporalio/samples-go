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
		workflow.Go(ctx, func(ctx workflow.Context) {
			defer wg.Done()
			// Use the cancelable callCtx for executing the operation.
			fut := c.ExecuteOperation(
				callCtx,
				service.HelloOperationName,
				service.HelloInput{Name: name, Language: lang},
				workflow.NexusOperationOptions{})
			var output service.HelloOutput
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
