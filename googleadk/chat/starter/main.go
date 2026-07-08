package main

import (
	"context"
	"log"
	"time"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"

	chat "github.com/temporalio/samples-go/googleadk/chat"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowID := "google-adk-chat_workflowID"
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: chat.TaskQueue,
	}

	// A small MaxTurns forces the continue-as-new boundary quickly for the demo.
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, chat.ChatWorkflow, chat.ChatInput{MaxTurns: 3})
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started chat workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	messages := []string{
		"Hi! My name is David.",
		"What's a fun fact about durable execution?",
	}
	for _, m := range messages {
		if err := c.SignalWorkflow(context.Background(), workflowID, "", chat.UserMessageSignalName, m); err != nil {
			log.Fatalln("Unable to send message signal", err)
		}
		log.Printf("Sent message: %q", m)
		// Give the agent time to answer before querying.
		time.Sleep(2 * time.Second)

		resp, err := c.QueryWorkflow(context.Background(), workflowID, "", chat.LatestAnswerQueryType)
		if err != nil {
			log.Fatalln("Unable to query latest answer", err)
		}
		var answer string
		if err := resp.Get(&answer); err != nil {
			log.Fatalln("Unable to decode query result", err)
		}
		log.Printf("Assistant: %q", answer)
	}

	log.Println("Done. The chat workflow keeps running (continuing-as-new to bound history); terminate it from the UI when finished.")
}
