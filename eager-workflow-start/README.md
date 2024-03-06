### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use) with eager workflow start enabled by a dynamic config value.
```
temporal server start-dev --dynamic-config-value system.enableEagerWorkflowStart=true
```
2) Run the following command to start the combined worker and example
```
go run eager-workflow-start/main.go
```

