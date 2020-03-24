This sample demonstrates how to implement a DSL workflow. In this sample, we provide 2 sample yaml files each defines a custom workflow that can be processed by this DSL workflow sample code.

Steps to run this sample:
1) You need a Temporal service running. See README.md for more details.
2) Run
```
go run dsl/worker/main.go
```
to start worker for dsl workflow.
3) Run 
```
go run dsl/starter/main.go
```
to submit start request for workflow defined in `workflow1.yaml` file.

Next:
1) You can run 
```
go run dsl/starter/main.go -dslConfig=dsl/workflow2.yaml
```
to see the result.
2) You can also write your own yaml config to play with it.
3) You can replace the dummy activities to your own real activities to build real workflow based on this simple DSL workflow.
