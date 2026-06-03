package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
	"go.temporal.io/sdk/contrib/workflowstreams"
	"go.temporal.io/sdk/converter"

	streams "github.com/temporalio/samples-go/workflowstreams"
)

// ANSI escapes to save the cursor position and to restore it while clearing
// everything below, so a retry can re-render the completion from scratch.
const (
	ansiSave            = "\x1b[s"
	ansiRestoreAndClear = "\x1b[u\x1b[J"
)

// Scenario 5: LLM token streaming. The workflow hosts the stream while an
// activity makes the streaming OpenAI call and republishes each token delta.
// On a retry the activity emits a RetryEvent and this subscriber rewinds the
// terminal and re-renders. Run ./workflowstreams/llmworker with OPENAI_API_KEY
// set before running this.
func main() {
	prompt := "In one short paragraph, explain what Temporal is."
	if len(os.Args) > 1 {
		prompt = os.Args[1]
	}

	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.Background()
	workflowID := "workflow-streams-llm-" + uuid.NewString()

	we, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: streams.LLMTaskQueue,
	}, streams.LLMWorkflow, streams.LLMInput{Prompt: prompt})
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	dc := converter.GetDefaultDataConverter()
	stream := workflowstreams.NewClient(c, workflowID, workflowstreams.Options{})
	defer func() { _ = stream.Close(ctx) }()

	fmt.Print(ansiSave)
	for item, err := range stream.Subscribe(ctx, workflowstreams.SubscribeOptions{
		Topics: []string{streams.TopicDelta, streams.TopicRetry, streams.TopicComplete},
	}) {
		if err != nil {
			log.Fatalln("subscribe:", err)
		}
		switch item.Topic {
		case streams.TopicRetry:
			var evt streams.RetryEvent
			if err := dc.FromPayload(item.Data, &evt); err != nil {
				log.Fatalln("decode retry:", err)
			}
			fmt.Print(ansiRestoreAndClear)
			fmt.Printf("[retry attempt %d] resetting output\n\n", evt.Attempt)
			fmt.Print(ansiSave)
		case streams.TopicDelta:
			var evt streams.TextDelta
			if err := dc.FromPayload(item.Data, &evt); err != nil {
				log.Fatalln("decode delta:", err)
			}
			fmt.Print(evt.Text)
		case streams.TopicComplete:
			fmt.Println()
			return
		}
	}
}
