# Child Workflow Type Validation Interceptor Sample

This sample shows how to make a worker interceptor that intercepts child workflow requests. 
The requests are validated using an activity. 
Most of the sample complexity is in creating a custom implementation of a ChildWorkflowResult.


### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run ./workflow-security-interceptor/worker
```
3) Run the following command to start the example
```
go run ./workflow-security-interceptor/starter
```
The expected output is workflow failure with the following message:
```
Child workflow type "UnallowedChildWorkflow" not allowed (type: not-allowed, retryable: true)
```

