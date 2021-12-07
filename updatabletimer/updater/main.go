package main

import (
	"context"
	"github.com/temporalio/samples-go/updatabletimer"
	"log"
	"time"

	"go.temporal.io/sdk/client"
)

// Signals updatable timer workflow to change wake-up time to 20 seconds from now.
func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	wakeUpTime := time.Now().Add(20 * time.Second)

	err = c.SignalWorkflow(context.Background(), updatabletimer.WorkflowID, "", updatabletimer.SignalType, wakeUpTime)
	if err != nil {
		log.Fatalln("Unable to signale workflow", err)
	}
	log.Println("Signaled workflow to update wake-up time",
		"WorkflowID", updatabletimer.WorkflowID, "WakeUpTime", wakeUpTime)
}
