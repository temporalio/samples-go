package main

import (
	"context"
	"flag"

	"go.temporal.io/temporal/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/recovery"
)

func main() {
	var workflowID string
	flag.StringVar(&workflowID, "w", "trip_workflow", "WorkflowID.")
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

	resp, err := c.QueryWorkflow(context.Background(), workflowID, "", recovery.QueryName)
	if err != nil {
		logger.Fatal("Unable to query workflow", zap.Error(err))
	}
	var result interface{}
	if err := resp.Get(&result); err != nil {
		logger.Error("Unable to decode query result", zap.Error(err))
	}
	logger.Info("Received query result", zap.Any("Result", result))

	// Close connection, clean up resources.
	_ = c.CloseConnection()
}
