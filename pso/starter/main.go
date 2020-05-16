package main

import (
	"context"
	"flag"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.uber.org/zap"

	"github.com/temporalio/temporal-go-samples/pso"
)

func main() {
	var functionName string
	flag.StringVar(&functionName, "f", "sphere", "One of [sphere, rosenbrock, griewank]")
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
	defer c.CloseConnection()

	workflowOptions := client.StartWorkflowOptions{
		ID:       "PSO_" + uuid.New(),
		TaskList: "pso",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, pso.PSOWorkflow, functionName)
	if err != nil {
		logger.Fatal("Unable to execute workflow", zap.Error(err))
	}
	logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))
}
