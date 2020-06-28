package main

import (
	"context"
	"flag"
	"io/ioutil"

	"github.com/pborman/uuid"
	"go.temporal.io/temporal/client"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/temporalio/temporal-go-samples/dsl"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	var dslConfig string
	flag.StringVar(&dslConfig, "dslConfig", "dsl/workflow1.yaml", "dslConfig specify the yaml file for the dsl workflow.")
	flag.Parse()

	data, err := ioutil.ReadFile(dslConfig)
	if err != nil {
		logger.Fatal("failed to load dsl config file", zap.Error(err))
	}
	var dslWorkflow dsl.Workflow
	if err := yaml.Unmarshal(data, &dslWorkflow); err != nil {
		logger.Fatal("failed to unmarshal dsl config", zap.Error(err))
	}

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		logger.Fatal("Unable to create client", zap.Error(err))
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        "dsl_" + uuid.New(),
		TaskQueue: "dsl",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, dsl.SimpleDSLWorkflow, dslWorkflow)
	if err != nil {
		logger.Fatal("Unable to execute workflow", zap.Error(err))
	}
	logger.Info("Started workflow", zap.String("WorkflowID", we.GetID()), zap.String("RunID", we.GetRunID()))

}
