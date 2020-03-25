package main

import (
	"context"
	"flag"

	"go.temporal.io/temporal/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/localactivity"
)

func main() {
	var workflowID, signalData string
	flag.StringVar(&workflowID, "w", "local_activity_workflow", "WorkflowID.")
	flag.StringVar(&signalData, "s", `{}`, "Signal data.")
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

	err = c.SignalWorkflow(context.Background(), workflowID, "", localactivity.SignalName, signalData)
	if err != nil {
		logger.Fatal("Unable to signal workflow", zap.Error(err))
	}

	// Close connection, clean up resources.
	_ = c.CloseConnection()
}
