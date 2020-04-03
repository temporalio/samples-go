module github.com/temporalio/temporal-go-samples

go 1.13

require (
	github.com/golang/mock v1.4.3
	github.com/pborman/uuid v1.2.0
	github.com/stretchr/testify v1.5.1
	go.temporal.io/temporal v0.20.4
	go.temporal.io/temporal-proto v0.20.11
	go.uber.org/zap v1.14.1
	gopkg.in/yaml.v2 v2.2.8
)

replace go.temporal.io/temporal v0.20.4 => ../temporal-go-sdk
