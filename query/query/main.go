package main

import (
	"context"
	"flag"

	"go.temporal.io/temporal/client"
	"go.uber.org/zap"
)

func main() {
	var workflowID, queryType string
	flag.StringVar(&workflowID, "w", "query_workflow", "WorkflowID.")
	flag.StringVar(&queryType, "t", "state", "Query type [state|__stack_trace].")
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
	defer c.CloseConnection()

	resp, err := c.QueryWorkflow(context.Background(), workflowID, "", queryType)
	if err != nil {
		logger.Fatal("Unable to query workflow", zap.Error(err))
	}
	var result interface{}
	if err := resp.Get(&result); err != nil {
		logger.Error("Unable to decode query result", zap.Error(err))
	}
	logger.Info("Received query result", zap.Any("Result", result))
}
