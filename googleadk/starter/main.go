package main

import (
	"context"
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"

	adk "github.com/temporalio/samples-go/googleadk"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "google-adk_workflowID",
		TaskQueue: adk.TaskQueue,
	}

	question := "What's the weather in San Francisco?"
	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, adk.AgentWorkflow, question)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var answer string
	if err := we.Get(context.Background(), &answer); err != nil {
		log.Fatalln("Unable to get workflow result", err)
	}
	log.Println("Agent answer:", answer)
}
