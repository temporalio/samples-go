module github.com/temporalio/samples-go

go 1.16

replace go.temporal.io/server => github.com/robholland/temporal v1.1.1-0.20210427090742-bc9685c643ea

require (
	github.com/HdrHistogram/hdrhistogram-go v0.9.0 // indirect
	github.com/golang/mock v1.5.0
	github.com/hashicorp/go-plugin v1.4.0
	github.com/m3db/prometheus_client_golang v0.8.1
	github.com/opentracing/opentracing-go v1.2.0
	github.com/pborman/uuid v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/uber-go/tally v3.3.17+incompatible
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	github.com/uber/jaeger-lib v2.4.0+incompatible // indirect
	go.temporal.io/api v1.4.1-0.20210420220407-6f00f7f98373
	go.temporal.io/sdk v1.6.0
	go.temporal.io/server v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
