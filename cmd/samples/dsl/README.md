This sample demonstrates how to implement a DSL workflow. In this sample, we provide 2 sample yaml files each defines a custom workflow that can be processed by this dsl workflow sample code.

Steps to run this sample:
1) You need a cadence service running. See cmd/samples/README.md for more details.
2) Run "./dsl -m worker" to start workers for dsl workflow.
3) Run "./dsl -dslConfig cmd/samples/dsl/workflow1.yaml" to submit start request for workflow defined in workflow1.yaml file.

Next:
1) You can replace the dslConfig to workflow2.yaml to see the result.
2) You can also write your own yaml config to play with it.
3) You can replace the dummy activities to your own real activities to build real workflow based on this simple dsl workflow.
