package common

import (
	"context"
	"fmt"
	"io/ioutil"

	"go.uber.org/cadence/worker"
	"go.uber.org/zap"

	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/client"
	"gopkg.in/yaml.v2"
)

const (
	configFile = "config/development.yaml"
)

type (
	// SampleHelper class for workflow sample helper.
	SampleHelper struct {
		Service workflowserviceclient.Interface
		Scope   tally.Scope
		Logger  *zap.Logger
		Config  Configuration
		Builder *WorkflowClientBuilder
	}

	// Configuration for running samples.
	Configuration struct {
		DomainName      string `yaml:"domain"`
		ServiceName     string `yaml:"service"`
		HostNameAndPort string `yaml:"host"`
	}
)

// SetupServiceConfig setup the config for the sample code run
func (h *SampleHelper) SetupServiceConfig() {
	if h.Service != nil {
		return
	}

	// Initialize developer config for running samples
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to log config file: %v, Error: %v", configFile, err))
	}

	if err := yaml.Unmarshal(configData, &h.Config); err != nil {
		panic(fmt.Sprintf("Error initializing configuration: %v", err))
	}

	// Initialize logger for running samples
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger.Info("Logger created.")
	h.Logger = logger
	h.Scope = tally.NoopScope
	h.Builder = NewBuilder(logger).
		SetHostPort(h.Config.HostNameAndPort).
		SetDomain(h.Config.DomainName).
		SetMetricsScope(h.Scope)
	service, err := h.Builder.BuildServiceClient()
	if err != nil {
		panic(err)
	}
	h.Service = service

	domainClient, _ := h.Builder.BuildCadenceDomainClient()
	_, err = domainClient.Describe(context.Background(), h.Config.DomainName)
	if err != nil {
		logger.Info("Domain doesn't exist", zap.String("Domain", h.Config.DomainName), zap.Error(err))
	} else {
		logger.Info("Domain succeesfully registered.", zap.String("Domain", h.Config.DomainName))
	}
}

// StartWorkflow starts a workflow
func (h *SampleHelper) StartWorkflow(options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) {
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}

	we, err := workflowClient.StartWorkflow(context.Background(), options, workflow, args...)
	if err != nil {
		h.Logger.Error("Failed to create workflow", zap.Error(err))
		panic("Failed to create workflow.")

	} else {
		h.Logger.Info("Started Workflow", zap.String("WorkflowID", we.ID), zap.String("RunID", we.RunID))
	}
}

// StartWorkers starts workflow worker and activity worker based on configured options.
func (h *SampleHelper) StartWorkers(domainName, groupName string, options worker.Options) {
	worker := worker.New(h.Service, domainName, groupName, options)
	err := worker.Start()
	if err != nil {
		h.Logger.Error("Failed to start workers.", zap.Error(err))
		panic("Failed to start workers")
	}
}

func (h *SampleHelper) QueryWorkflow(workflowID, runID, queryType string, args ...interface{}) {
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}

	resp, err := workflowClient.QueryWorkflow(context.Background(), workflowID, runID, queryType, args...)
	if err != nil {
		h.Logger.Error("Failed to query workflow", zap.Error(err))
		panic("Failed to query workflow.")
	}
	var result interface{}
	if err := resp.Get(&result); err != nil {
		h.Logger.Error("Failed to decode query result", zap.Error(err))
	}
	h.Logger.Info("Received query result", zap.Any("Result", result))
}

func (h *SampleHelper) SignalWorkflow(workflowID, signal string, data interface{}) {
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}

	err = workflowClient.SignalWorkflow(context.Background(), workflowID, "", signal, data)
	if err != nil {
		h.Logger.Error("Failed to signal workflow", zap.Error(err))
		panic("Failed to signal workflow.")
	}
}

func (h *SampleHelper) CancelWorkflow(workflowID string) {
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}

	err = workflowClient.CancelWorkflow(context.Background(), workflowID, "")
	if err != nil {
		h.Logger.Error("Failed to cancel workflow", zap.Error(err))
		panic("Failed to cancel workflow.")
	}
}
