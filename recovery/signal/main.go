package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"

	"go.temporal.io/sdk/client"

	"github.com/temporalio/samples-go/recovery"
)

func main() {
	var workflowID, signal string
	flag.StringVar(&workflowID, "w", "trip_workflow", "WorkflowID.")
	flag.StringVar(&signal, "s", `{}`, "Signal data.")
	flag.Parse()

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	var tripEvent recovery.TripEvent
	if err := json.Unmarshal([]byte(signal), &tripEvent); err != nil {
		log.Fatalln("Unable to unmarshal signal input parameters", err)
	}

	err = c.SignalWorkflow(context.Background(), workflowID, "", recovery.TripSignalName, tripEvent)
	if err != nil {
		log.Fatalln("Unable to signal workflow", err)
	}
}
