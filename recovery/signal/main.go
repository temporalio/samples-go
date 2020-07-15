package main

import (
	"context"
	"encoding/json"
	"flag"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/recovery"
)

func main() {
	var workflowID, signal string
	flag.StringVar(&workflowID, "w", "trip_workflow", "WorkflowID.")
	flag.StringVar(&signal, "s", `{}`, "Signal data.")
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer c.Close()

	var tripEvent recovery.TripEvent
	if err := json.Unmarshal([]byte(signal), &tripEvent); err != nil {
		logger.Fatal("Unable to unmarshal signal input parameters", zap.Error(err))
	}

	err = c.SignalWorkflow(context.Background(), workflowID, "", recovery.TripSignalName, tripEvent)
	if err != nil {
		logger.Fatal("Unable to signal workflow", zap.Error(err))
	}
}
