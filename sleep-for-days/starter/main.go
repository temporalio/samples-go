package main

import (
	"context"
	"log"

	sleepfordays "github.com/temporalio/samples-go/sleep-for-days"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/envconfig"
)

func main() {
	c, err := client.Dial(envconfig.MustLoadDefaultClientOptions())
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "sleep-for-days",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, sleepfordays.SleepForDaysWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started sleep-for-days workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion (will run indefinitely until it receives a signal)
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
	log.Println("Workflow result:", result)
}
