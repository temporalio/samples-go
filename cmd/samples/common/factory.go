package common

import (
	"errors"

	"github.com/uber-go/tally"
	"go.temporal.io/temporal-proto/workflowservice"
	"go.temporal.io/temporal/client"
	"go.temporal.io/temporal/encoded"
	"go.temporal.io/temporal/workflow"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// WorkflowClientBuilder build client to cadence service
type WorkflowClientBuilder struct {
	hostPort       string
	domain         string
	clientIdentity string
	metricsScope   tally.Scope
	Logger         *zap.Logger
	ctxProps       []workflow.ContextPropagator
	dataConverter  encoded.DataConverter
}

// NewBuilder creates a new WorkflowClientBuilder
func NewBuilder(logger *zap.Logger) *WorkflowClientBuilder {
	return &WorkflowClientBuilder{
		Logger: logger,
	}
}

// SetHostPort sets the hostport for the builder
func (b *WorkflowClientBuilder) SetHostPort(hostport string) *WorkflowClientBuilder {
	b.hostPort = hostport
	return b
}

// SetDomain sets the domain for the builder
func (b *WorkflowClientBuilder) SetDomain(domain string) *WorkflowClientBuilder {
	b.domain = domain
	return b
}

// SetClientIdentity sets the identity for the builder
func (b *WorkflowClientBuilder) SetClientIdentity(identity string) *WorkflowClientBuilder {
	b.clientIdentity = identity
	return b
}

// SetMetricsScope sets the metrics scope for the builder
func (b *WorkflowClientBuilder) SetMetricsScope(metricsScope tally.Scope) *WorkflowClientBuilder {
	b.metricsScope = metricsScope
	return b
}

// SetContextPropagators sets the context propagators for the builder
func (b *WorkflowClientBuilder) SetContextPropagators(ctxProps []workflow.ContextPropagator) *WorkflowClientBuilder {
	b.ctxProps = ctxProps
	return b
}

// SetDataConverter sets the data converter for the builder
func (b *WorkflowClientBuilder) SetDataConverter(dataConverter encoded.DataConverter) *WorkflowClientBuilder {
	b.dataConverter = dataConverter
	return b
}

// BuildCadenceClient builds a client to cadence service
func (b *WorkflowClientBuilder) BuildCadenceClient() (client.Client, error) {
	service, err := b.BuildServiceClient()
	if err != nil {
		return nil, err
	}

	return client.NewClient(
		service, b.domain, &client.Options{Identity: b.clientIdentity, MetricsScope: b.metricsScope, DataConverter: b.dataConverter, ContextPropagators: b.ctxProps}), nil
}

// BuildCadenceDomainClient builds a domain client to cadence service
func (b *WorkflowClientBuilder) BuildCadenceDomainClient() (client.DomainClient, error) {
	service, err := b.BuildServiceClient()
	if err != nil {
		return nil, err
	}

	return client.NewDomainClient(
		service, &client.Options{Identity: b.clientIdentity, MetricsScope: b.metricsScope, ContextPropagators: b.ctxProps}), nil
}

// BuildServiceClient builds a rpc service client to cadence service
func (b *WorkflowClientBuilder) BuildServiceClient() (workflowservice.WorkflowServiceClient, error) {
	if len(b.hostPort) == 0 {
		return nil, errors.New("HostPort is empty")
	}

	connection, err := grpc.Dial(b.hostPort, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	// TODO: close connection somewhere
	return workflowservice.NewWorkflowServiceClient(connection), nil
}
