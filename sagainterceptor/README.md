### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run sagainterceptor/worker/main.go
```
3) Run the following command to start the example
```
go run sagainterceptor/start/main.go
```


Workflow definition has no compensations, it is another style to implement saga pattern compared to saga/workflow.go.

Based on https://github.com/temporalio/money-transfer-project-template-go
