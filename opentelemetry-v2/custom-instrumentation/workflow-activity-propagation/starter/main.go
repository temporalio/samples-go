package main

import (
	"context"
	"fmt"
	"log"

	workflowactivity "github.com/temporalio/samples-go/opentelemetry-v2/custom-instrumentation/workflow-activity-propagation"
	"go.temporal.io/sdk/client"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	ctx := context.Background()
	c, err := client.Dial(client.Options{})
	if err != nil {
		return fmt.Errorf("unable to create client: %w", err)
	}
	defer c.Close()

	we, err := c.ExecuteWorkflow(
		ctx,
		client.StartWorkflowOptions{TaskQueue: workflowactivity.TaskQueueName},
		workflowactivity.Workflow,
		"Temporal",
	)
	if err != nil {
		return fmt.Errorf("unable to execute workflow: %w", err)
	}

	var result string
	if err := we.Get(ctx, &result); err != nil {
		return fmt.Errorf("unable to get workflow result: %w", err)
	}

	log.Println("Workflow result:", result)
	return nil
}
