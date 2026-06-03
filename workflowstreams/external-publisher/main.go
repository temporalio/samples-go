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

var headlines = []string{
	"markets open higher",
	"new bridge opens downtown",
	"local team wins championship",
}

// doneHeadline is the sentinel the publisher sends last so the subscriber knows
// to stop.
const doneHeadline = "-- end of feed --"

// Scenario 3: external (non-activity) publisher. The hub workflow does no work of
// its own; it just hosts the stream. A separate process publishes news into it
// using the same client factory used to subscribe, then signals the workflow to
// close. Here the publisher and a subscriber run as two goroutines.
func main() {
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()
	workflowID := "workflow-streams-hub-" + uuid.NewString()

	we, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: streams.TaskQueue,
	}, streams.HubWorkflow, streams.HubInput{HubID: "newsroom"})
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	var wg sync.WaitGroup
	wg.Add(2)

	// Subscriber goroutine.
	go func() {
		defer wg.Done()
		dc := converter.GetDefaultDataConverter()
		stream := workflowstreams.NewClient(c, workflowID, workflowstreams.Options{})
		defer func() { _ = stream.Close(ctx) }()
		for item, err := range stream.Topic(streams.TopicNews).Subscribe(ctx, 0) {
			if err != nil {
				log.Fatalln("subscribe:", err)
			}
			var evt streams.NewsEvent
			if err := dc.FromPayload(item.Data, &evt); err != nil {
				log.Fatalln("decode news:", err)
			}
			if evt.Headline == doneHeadline {
				return
			}
			fmt.Printf("[subscriber] %s\n", evt.Headline)
		}
	}()

	// Publisher goroutine.
	go func() {
		defer wg.Done()
		producer := workflowstreams.NewClient(c, workflowID, workflowstreams.Options{})
		defer func() { _ = producer.Close(ctx) }()
		news := producer.Topic(streams.TopicNews)
		for _, headline := range headlines {
			news.Publish(streams.NewsEvent{Headline: headline}, false)
			fmt.Printf("[publisher]  sent: %s\n", headline)
			time.Sleep(500 * time.Millisecond)
		}
		// Force-flush the sentinel and wait for the server to confirm delivery.
		news.Publish(streams.NewsEvent{Headline: doneHeadline}, true)
		if err := producer.Flush(ctx); err != nil {
			log.Fatalln("flush:", err)
		}
		if err := c.SignalWorkflow(ctx, workflowID, "", streams.CloseSignal, nil); err != nil {
			log.Fatalln("signal close:", err)
		}
		fmt.Println("[publisher]  signaled close")
	}()

	wg.Wait()
}
