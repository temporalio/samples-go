package main

import (
	"context"
	"log"

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
		// Send the message as an Update and get the agent's answer back on the same
		// call — no signal + query polling.
		handle, err := c.UpdateWorkflow(context.Background(), client.UpdateWorkflowOptions{
			WorkflowID:   workflowID,
			UpdateName:   chat.SendMessageUpdateName,
			WaitForStage: client.WorkflowUpdateStageCompleted,
			Args:         []interface{}{m},
		})
		if err != nil {
			log.Fatalln("Unable to send message update", err)
		}
		var answer string
		if err := handle.Get(context.Background(), &answer); err != nil {
			log.Fatalln("Unable to get update result", err)
		}
		log.Printf("You: %q", m)
		log.Printf("Assistant: %q", answer)
	}

	log.Println("Done. The chat workflow keeps running (continuing-as-new to bound history); terminate it from the UI when finished.")
}
