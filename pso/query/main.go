package main

import (
	"context"
	"flag"

	"go.temporal.io/sdk/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/pso"
)

func main() {
	var workflowID, runID, queryType string
	flag.StringVar(&workflowID, "w", "", "WorkflowID")
	flag.StringVar(&runID, "r", "", "RunID")
	flag.StringVar(&queryType, "t", "__stack_trace", "Query type is one of [__stack_trace, child, __open_sessions]")
	flag.Parse()

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort:      client.DefaultHostPort,
		DataConverter: pso.NewJSONDataConverter(),
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer c.Close()

	resp, err := c.QueryWorkflow(context.Background(), workflowID, runID, queryType)
	if err != nil {
		logger.Fatal("Unable to query workflow", zap.Error(err))
	}
	var result interface{}
	if err := resp.Get(&result); err != nil {
		logger.Error("Unable to decode query result", zap.Error(err))
	}
	logger.Info("Received query result", zap.Any("Result", result))
}
