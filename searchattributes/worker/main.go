package main

import (
	"context"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/searchattributes"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
		Logger:   logger,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}
	defer c.Close()

	ctx := context.WithValue(context.Background(), searchattributes.TemporalClientKey, c)

	w := worker.New(c, "search-attributes", worker.Options{
		BackgroundActivityContext: ctx,
	})

	w.RegisterWorkflow(searchattributes.SearchAttributesWorkflow)
	w.RegisterActivity(searchattributes.ListExecutions)

	err = w.Run()
	if err != nil {
		logger.Fatal("Unable to start worker", zap.Error(err))
	}
}
