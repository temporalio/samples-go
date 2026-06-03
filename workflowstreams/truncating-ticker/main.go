package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/contrib/workflowstreams"
	"go.temporal.io/sdk/converter"

	streams "github.com/temporalio/samples-go/workflowstreams"
)

const (
	tickCount = 30
	// keepLast bounds the workflow's log to its most recent entries;
	// truncateEvery controls how often it truncates. They are deliberately small
	// so the early offsets are dropped quickly.
	keepLast      = 5
	truncateEvery = 5
	// staleOffset is an early offset the late subscriber deliberately resumes
	// from. By the time it subscribes, the workflow has truncated past it.
	staleOffset = 1
)

// Scenario 4: bounded log via truncation. The ticker workflow periodically
// truncates old entries to bound its history, trading complete history for a
// bounded log. A "fast" subscriber that reads from the start keeps up and sees
// every tick. A "late" subscriber that joins after truncation and resumes from a
// stale offset is fast-forwarded to the current base offset — it cannot see the
// truncated ticks.
func main() {
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()
	workflowID := "workflow-streams-ticker-" + uuid.NewString()

	we, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: streams.TaskQueue,
	}, streams.TickerWorkflow, streams.TickerInput{
		Count:         tickCount,
		KeepLast:      keepLast,
		TruncateEvery: truncateEvery,
	})
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	lastN := tickCount - 1

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		fastSubscriber(ctx, c, workflowID, lastN)
	}()
	go func() {
		defer wg.Done()
		lateSubscriber(ctx, c, workflowID, lastN)
	}()

	wg.Wait()
}

// fastSubscriber reads from the beginning and keeps up with every tick.
func fastSubscriber(ctx context.Context, c client.Client, workflowID string, lastN int) {
	dc := converter.GetDefaultDataConverter()
	stream := workflowstreams.NewClient(c, workflowID, workflowstreams.Options{})
	defer func() { _ = stream.Close(ctx) }()

	for item, err := range stream.Topic(streams.TopicTick).Subscribe(ctx, 0) {
		if err != nil {
			log.Fatalln("fast subscribe:", err)
		}
		var evt streams.TickEvent
		if err := dc.FromPayload(item.Data, &evt); err != nil {
			log.Fatalln("decode tick:", err)
		}
		fmt.Printf("[fast] offset=%3d  n=%d\n", item.Offset, evt.N)
		if evt.N == lastN {
			return
		}
	}
}

// lateSubscriber waits until the workflow has truncated past staleOffset, then
// resumes from that (now-truncated) offset. The stream fast-forwards it to the
// current base offset, so its first item necessarily skips the truncated ticks.
func lateSubscriber(ctx context.Context, c client.Client, workflowID string, lastN int) {
	dc := converter.GetDefaultDataConverter()
	stream := workflowstreams.NewClient(c, workflowID, workflowstreams.Options{})
	defer func() { _ = stream.Close(ctx) }()

	// Wait until the workflow has truncated past staleOffset. The first
	// truncation only fires once published reaches the first multiple of
	// truncateEvery greater than keepLast; after it, the base offset is
	// published-keepLast. Waiting until head passes that point guarantees the
	// base has advanced beyond staleOffset.
	firstTruncate := ((keepLast / truncateEvery) + 1) * truncateEvery
	for {
		head, err := stream.GetOffset(ctx)
		if err != nil {
			log.Fatalln("get offset:", err)
		}
		if int(head) > firstTruncate {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	first := true
	for item, err := range stream.Topic(streams.TopicTick).Subscribe(ctx, staleOffset) {
		if err != nil {
			log.Fatalln("late subscribe:", err)
		}
		var evt streams.TickEvent
		if err := dc.FromPayload(item.Data, &evt); err != nil {
			log.Fatalln("decode tick:", err)
		}
		if first {
			if item.Offset > staleOffset {
				fmt.Printf("[late] requested offset %d but it was truncated; fast-forwarded to offset %d (skipped %d tick(s))\n",
					staleOffset, item.Offset, item.Offset-staleOffset)
			}
			first = false
		}
		fmt.Printf("[late] offset=%3d  n=%d\n", item.Offset, evt.N)
		if evt.N == lastN {
			return
		}
	}
}
