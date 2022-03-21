module github.com/temporalio/samples-go

go 1.16

replace github.com/cactus/go-statsd-client => github.com/cactus/go-statsd-client v3.2.1+incompatible

require (
	github.com/golang/mock v1.6.0
	github.com/golang/snappy v0.0.4
	github.com/hashicorp/go-plugin v1.4.3
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pborman/uuid v1.2.1
	github.com/prometheus/client_golang v1.11.0
	github.com/stretchr/testify v1.7.0
	github.com/uber-go/tally/v4 v4.1.1
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	go.temporal.io/api v1.7.1-0.20220223032354-6e6fe738916a
	go.temporal.io/sdk v1.14.0
	go.temporal.io/sdk/contrib/opentracing v0.1.0
	go.temporal.io/sdk/contrib/tally v0.1.0
	// TODO(cretz): Remove when tagged
	go.temporal.io/sdk/contrib/tools/workflowcheck v0.0.0-00010101000000-000000000000
	go.temporal.io/server v1.14.1
	go.uber.org/zap v1.19.1
	google.golang.org/grpc v1.44.0
	gopkg.in/check.v1 v1.0.0-20200902074654-038fdea0a05b // indirect
	gopkg.in/square/go-jose.v2 v2.6.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

// TODO(cretz): Remove when tagged (can remove entire dependency)
replace go.temporal.io/sdk/contrib/tools/workflowcheck => github.com/cretz/temporal-sdk-go/contrib/tools/workflowcheck v0.0.0-20220121193620-36ed5f1888d9
