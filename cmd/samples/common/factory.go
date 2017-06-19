package common

import (
	"errors"

	"go.uber.org/cadence"
	m "go.uber.org/cadence/.gen/go/cadence"

	"github.com/uber-go/tally"
	"github.com/uber/tchannel-go"
	"github.com/uber/tchannel-go/thrift"
)

const (
	cadenceClientName      = "cadence-client"
	cadenceFrontendService = "cadence-frontend"
)

// WorkflowClientBuilder build client to cadence service
type WorkflowClientBuilder struct {
	tchanClient    thrift.TChanClient
	hostPort       string
	domain         string
	clientIdentity string
	metricsScope   tally.Scope
}

// NewBuilder creates a new WorkflowClientBuilder
func NewBuilder() *WorkflowClientBuilder {
	return &WorkflowClientBuilder{}
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

// BuildCadenceClient builds a client to cadence service
func (b *WorkflowClientBuilder) BuildCadenceClient() (cadence.Client, error) {
	service, err := b.BuildServiceClient()
	if err != nil {
		return nil, err
	}

	return cadence.NewClient(
		service, b.domain, &cadence.ClientOptions{Identity: b.clientIdentity, MetricsScope: b.metricsScope}), nil
}

// BuildCadenceDomainClient builds a domain client to cadence service
func (b *WorkflowClientBuilder) BuildCadenceDomainClient() (cadence.DomainClient, error) {
	service, err := b.BuildServiceClient()
	if err != nil {
		return nil, err
	}

	return cadence.NewDomainClient(
		service, &cadence.ClientOptions{Identity: b.clientIdentity, MetricsScope: b.metricsScope}), nil
}

// BuildServiceClient builds a thrift service client to cadence service
func (b *WorkflowClientBuilder) BuildServiceClient() (m.TChanWorkflowService, error) {
	if err := b.build(); err != nil {
		return nil, err
	}

	return m.NewTChanWorkflowServiceClient(b.tchanClient), nil
}

func (b *WorkflowClientBuilder) build() error {
	if b.tchanClient != nil {
		return nil
	}
	if len(b.hostPort) == 0  {
		return errors.New("HostPort must be valid")
	}

	tchan, err := tchannel.NewChannel(cadenceClientName, nil)
	if err != nil {
		return err
	}

	opts := &thrift.ClientOptions{HostPort: b.hostPort}
	b.tchanClient = thrift.NewClient(tchan, cadenceFrontendService, opts)
	return nil
}