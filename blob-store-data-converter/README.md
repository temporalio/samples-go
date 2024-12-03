# Blobstore DataConverter
This sample demonstrates how to use the DataConverter to store large payloads greater than a certain size 
in a blobstore and passes the object path around in the Temporal Event History.

The payload size limit is set in [codec.go: `payloadSizeLimit`](./codec.go#L20).

It relies on the use of context propagation to pass blobstore config metadata, like object path prefixes.

In this example, we prefix all object paths with a `tenantID` to better object lifecycle in the blobstore.

> [!NOTE]
> The time it takes to encode/decode payloads is counted in the `StartWorkflowOptions.WorkflowTaskTimeout`,
> which has a [absolute max of 2 minutes](https://github.com/temporalio/temporal/blob/2a0f6b238f6cdab768098194436b0dda453c8064/common/constants.go#L68). 

> [!WARNING]
> As of `Temporal UI v2.33.1` (`Temporal v1.25.2`), **does not** have the ability to send context headers.
> This means that Workflow Start, Signal, Queries, etc. from the UI/CLI will pass payloads to the codec-server but the 
> worker needs to handle a missing context propagation header.
> 
> In this sample when the header is missing, we use a default of `DefaultPropagatedValues()`,
> see [propagator.go: `missingHeaderContextPropagationKeyError`](./propagator.go#L66).
> 
> This allows this sample to still work with the UI/CLI. This maybe not suitable depending on your requirements. 


### Steps to run this sample:
1. Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use)
2. Run the following command to start the worker
    ```
    go run worker/main.go
    ```
3. Run the following command to start the example
    ```
    go run starter/main.go
    ```
4. Open the Temporal Web UI and observe the following:
   - Workflow Input and Activity Input values will be in plain text
   - Activity Result and Workflow Result will be an object path
5. Run the following command to start the remote codec server
    ```
    go run ./codec-server
    ```
6. Open the Temporal Web UI and observe the workflow execution, all payloads will now be fully expanded.
7. You can use the Temporal CLI as well
    ```
    # payloads will be obfuscated object paths
    temporal workflow show --env local -w WORKFLOW_ID

    # payloads will be fully rendered json
    temporal --codec-endpoint 'http://localhost:8081/' workflow show -w WORKFLOW_ID
    ``````

Note: Please see the [codec-server](../codec-server/) sample for a more complete example of a codec server which provides oauth.
