package common

import (
	"context"
	"fmt"
	"io/ioutil"

	s "go.uber.org/cadence/.gen/go/shared"
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

var domainCreated bool

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

	if domainCreated {
		return
	}
	domainClient, _ := h.Builder.BuildCadenceDomainClient()
	description := "domain for cadence sample code"
	var retention int32 = 3
	request := &s.RegisterDomainRequest{
		Name:                                   &h.Config.DomainName,
		Description:                            &description,
		WorkflowExecutionRetentionPeriodInDays: &retention}

	err = domainClient.Register(context.Background(), request)
	if err != nil {
		if _, ok := err.(*s.DomainAlreadyExistsError); !ok {
			panic(err)
		}
		logger.Info("Domain already registered.", zap.String("Domain", h.Config.DomainName))
	} else {
		logger.Info("Domain succeesfully registered.", zap.String("Domain", h.Config.DomainName))
	}
	domainCreated = true
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
	worker := worker.NewWorker(h.Service, domainName, groupName, options)
	err := worker.Start()
	if err != nil {
		h.Logger.Error("Failed to start workers.", zap.Error(err))
		panic("Failed to start workers")
	}
}
