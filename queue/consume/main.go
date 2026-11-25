package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/temporalio/samples-go/queue"
)

// helper workflow to cancel resource request
func cancelResourceRequest(ctx workflow.Context, resourcePoolWorkflowID string, targetWorkflowID string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("start cancel resource request workflow",
		"resourcePoolWorkflowID", resourcePoolWorkflowID,
		"targetWorkflowID", targetWorkflowID)

	// create resource pool workflow execution reference
	execution := workflow.Execution{
		ID: resourcePoolWorkflowID,
	}

	// send cancel command
	err := queue.CancelResourceRequest(ctx, execution, targetWorkflowID)
	if err != nil {
		logger.Error("failed to cancel resource request", "Error", err)
		return err
	}

	logger.Info("successfully canceled resource request")
	return nil
}

// helper workflow to update resource pool size
func updateResourcePoolSize(ctx workflow.Context, resourcePoolWorkflowID string, newSize int) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("start update resource pool size workflow",
		"resourcePoolWorkflowID", resourcePoolWorkflowID,
		"newSize", newSize)

	// create resource pool workflow execution reference
	execution := workflow.Execution{
		ID: resourcePoolWorkflowID,
	}

	// send update command
	err := queue.UpdateResourcePool(ctx, execution, newSize)
	if err != nil {
		logger.Error("failed to update resource pool size", "Error", err)
		return err
	}

	logger.Info("successfully updated resource pool size")
	return nil
}

func main() {
	// create temporal client
	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233",
	})
	if err != nil {
		log.Fatalln("can't create client", err)
	}
	defer c.Close()
	// create sample workflow worker - also need to set context
	sampleWorker := worker.New(c, "queue-sample", worker.Options{
		BackgroundActivityContext: context.WithValue(context.Background(), queue.ClientContextKey, c),
	})

	// register sample workflow
	sampleWorker.RegisterWorkflow(queue.SampleWorkflowWithResourcePool)
	// register resource pool management workflow
	sampleWorker.RegisterWorkflow(cancelResourceRequest)
	sampleWorker.RegisterWorkflow(updateResourcePoolSize)

	// start all workers
	workerErr := make(chan error, 1)

	go func() {
		workerErr <- sampleWorker.Run(worker.InterruptCh())
	}()

	// wait for any worker to fail
	err = <-workerErr
	if err != nil {
		log.Fatalln("worker run failed", err)
	}
}
