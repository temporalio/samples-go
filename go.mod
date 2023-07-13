module github.com/temporalio/samples-go

go 1.16

replace github.com/cactus/go-statsd-client => github.com/cactus/go-statsd-client v3.2.1+incompatible

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/benbjohnson/clock v1.3.0 // indirect
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.3.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pborman/uuid v1.2.1
	github.com/prometheus/client_golang v1.12.1
	github.com/stretchr/testify v1.8.4
	github.com/twmb/murmur3 v1.1.6 // indirect
	github.com/uber-go/tally/v4 v4.1.1
	github.com/uber/jaeger-client-go v2.29.1+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	go.opentelemetry.io/otel v1.16.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.16.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.16.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.16.0
	go.opentelemetry.io/otel/sdk v1.16.0
	go.opentelemetry.io/otel/trace v1.16.0
	go.opentelemetry.io/proto/otlp v0.20.0 // indirect
	go.temporal.io/api v1.23.0
	go.temporal.io/sdk v1.23.1
	go.temporal.io/sdk/contrib/opentelemetry v0.2.0
	go.temporal.io/sdk/contrib/opentracing v0.1.0
	go.temporal.io/sdk/contrib/tally v0.2.0
	go.temporal.io/sdk/contrib/tools/workflowcheck v0.0.0-20230612164027-11c2cb9e7d2d
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.20.0
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/tools v0.9.3 // indirect
	google.golang.org/genproto v0.0.0-20230711160842-782d3b101e98 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20230711160842-782d3b101e98 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20230711160842-782d3b101e98 // indirect
	gopkg.in/yaml.v3 v3.0.1
)
