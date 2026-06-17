package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/contrib/workflowstreams"
	"go.temporal.io/sdk/converter"

	streams "github.com/temporalio/samples-go/workflowstreams"
)

// Scenario 1: basic publish/subscribe. Start an order workflow that publishes
// status events itself and runs an activity that publishes progress events to
// the same stream, then subscribe to both topics until the order completes.
func main() {
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()
	workflowID := "workflow-streams-order-" + uuid.NewString()

	we, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: streams.TaskQueue,
	}, streams.OrderWorkflow, streams.OrderInput{OrderID: "order-42"})
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	dc := converter.GetDefaultDataConverter()
	stream := workflowstreams.NewClient(c, workflowID, workflowstreams.Options{})
	defer func() { _ = stream.Close(ctx) }()

	for item, err := range stream.Subscribe(ctx, workflowstreams.SubscribeOptions{
		Topics: []string{streams.TopicStatus, streams.TopicProgress},
	}) {
		if err != nil {
			log.Fatalln("subscribe:", err)
		}
		switch item.Topic {
		case streams.TopicStatus:
			var evt streams.StatusEvent
			if err := dc.FromPayload(item.Data, &evt); err != nil {
				log.Fatalln("decode status:", err)
			}
			fmt.Printf("[status]   %s: order=%s\n", evt.Kind, evt.OrderID)
			if evt.Kind == "complete" {
				return
			}
		case streams.TopicProgress:
			var evt streams.ProgressEvent
			if err := dc.FromPayload(item.Data, &evt); err != nil {
				log.Fatalln("decode progress:", err)
			}
			fmt.Printf("[progress] %s\n", evt.Message)
		}
	}
}
