### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the remote codec server
```
go run ./codec-server
```
3) Run the following command to start the worker
```
go run worker/main.go
```
4) Run the following command to start the example
```
go run starter/main.go
```
5) Run the following command and see the payloads cannot be decoded
```
temporal workflow show --workflow-id snappycompress_workflowID
```
6) Run the following command and see the decoded payloads
```
temporal workflow show --workflow-id snappycompress_workflowID --codec-endpoint 'http://localhost:8081/'
```

Note: The codec server provided in this sample does not support decoding payloads for the Temporal Web UI, only Temporal CLI.
Please see the [codec-server](../codec-server/) sample for a more complete example of a codec server which provides UI decoding and oauth.
