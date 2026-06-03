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

// phase1Events is how many events the first subscriber reads before disconnecting.
const phase1Events = 2

// Scenario 2: reconnecting subscriber. A subscriber reads a few events, drops
// its connection, then a brand-new client resumes from the saved offset without
// missing events or seeing duplicates — because the events are durable in
// workflow history, not just held in memory.
func main() {
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()
	workflowID := "workflow-streams-pipeline-" + uuid.NewString()

	we, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: streams.TaskQueue,
	}, streams.PipelineWorkflow, streams.PipelineInput{PipelineID: "pipeline-7"})
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	dc := converter.GetDefaultDataConverter()

	// next is the offset to resume from: one past the last item we consumed.
	var next int64

	// Phase 1: connect, read a couple of events, remember our position, disconnect.
	fmt.Println("--- phase 1: initial subscriber ---")
	stream1 := workflowstreams.NewClient(c, workflowID, workflowstreams.Options{})
	seen := 0
	for item, err := range stream1.Topic(streams.TopicStatus).Subscribe(ctx, 0) {
		if err != nil {
			log.Fatalln("subscribe:", err)
		}
		var evt streams.StageEvent
		if err := dc.FromPayload(item.Data, &evt); err != nil {
			log.Fatalln("decode stage:", err)
		}
		next = item.Offset + 1
		fmt.Printf("offset=%d  stage=%s\n", item.Offset, evt.Stage)
		seen++
		if seen >= phase1Events {
			break
		}
	}
	_ = stream1.Close(ctx)

	fmt.Printf("--- disconnected; will resume from offset %d ---\n", next)

	// Phase 2: a new client resumes from the saved offset until the pipeline completes.
	fmt.Println("--- phase 2: reconnected subscriber ---")
	stream2 := workflowstreams.NewClient(c, workflowID, workflowstreams.Options{})
	defer func() { _ = stream2.Close(ctx) }()
	for item, err := range stream2.Topic(streams.TopicStatus).Subscribe(ctx, next) {
		if err != nil {
			log.Fatalln("subscribe:", err)
		}
		var evt streams.StageEvent
		if err := dc.FromPayload(item.Data, &evt); err != nil {
			log.Fatalln("decode stage:", err)
		}
		fmt.Printf("offset=%d  stage=%s\n", item.Offset, evt.Stage)
		if evt.Stage == "complete" {
			return
		}
	}
}
