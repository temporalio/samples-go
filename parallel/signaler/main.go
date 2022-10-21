package main

import (
	"context"
	"log"
	"os"

	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	args := os.Args[1:]

	if len(args) != 3 {
		log.Fatal("Call this program with the following arguments <workflow-id> <run-id> <signal-name>")
	}

	workflowID := args[0]
	runID := args[1]
	signalName := args[2]

	if signalName != "branch1" && signalName != "branch2" {
		log.Fatal("<signal-name> input must be branch1 or branch2")
	}
	ctx := context.Background()
	err = c.SignalWorkflow(ctx, workflowID, runID, signalName, "dummy-input")
	if err != nil {
		log.Fatalf("Failed signaling with error %v", err)
	}
	log.Println("Signaled workflow")
}
