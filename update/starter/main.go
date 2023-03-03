package main

import (
	"context"
	"log"

	"github.com/temporalio/samples-go/update"
	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "update-workflow-ID",
		TaskQueue: "update",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, update.Counter)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	mult := 1
	for i := 0; i < 10; i++ {
		addend := mult * i
		mult *= -1 // flip addend between negative and positive for each iteration
		handle, err := c.UpdateWorkflow(context.Background(), we.GetID(), we.GetRunID(), update.FetchAndAdd, addend)
		if err != nil {
			log.Fatal("error issuing update request", err)
		}
		var result int
		err = handle.Get(context.Background(), &result)
		if err != nil {
			log.Printf("fetch_and_add with addend %v failed: %v", addend, err)
		} else {
			log.Printf("fetch_and_add with addend %v succeeded: %v", addend, result)
		}
	}

	if err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), update.Done, nil); err != nil {
		log.Fatalf("failed to send %q signal to workflow: %v", update.Done, err)
	}
	var wfresult int
	if err = we.Get(context.Background(), &wfresult); err != nil {
		log.Fatalf("unable get workflow result: %v", err)
	}
	log.Println("workflow result:", wfresult)
}
