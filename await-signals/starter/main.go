package main

import (
	"context"
	"fmt"
	await_signals "github.com/temporalio/samples-go/await-signals"
	"log"
	"math/rand"
	"time"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "await_signals_" + uuid.New(),
		TaskQueue: "await_signals",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, await_signals.AwaitSignalsWorkflow)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	log.Println("Sending signals")
	signals := []int{1, 2, 3}
	// Send signals in random order
	rand.Shuffle(len(signals), func(i, j int) { signals[i], signals[j] = signals[j], signals[i] })
	for _, signal := range signals {
		signalName := fmt.Sprintf("Signal%d", signal)
		err = c.SignalWorkflow(context.Background(), we.GetID(), we.GetRunID(), signalName, nil)
		if err != nil {
			log.Fatalln("Unable to signals workflow", err)
		}
		log.Println("Sent " + signalName)
		time.Sleep(2 * time.Second)
	}

}
