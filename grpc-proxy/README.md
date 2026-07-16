### Sample overview:

This sample shows how to compress payloads using a GRPC proxy. This allows managing the codec(s) centrally rather than needing to configure each client/worker.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) `grpc-proxy/` is a separate Go module. Run all commands below from the `grpc-proxy/` directory:
```
cd grpc-proxy
```
3) Run the following command to start the GRPC proxy listening on port 8081
```
go run ./proxy-server
```
4) Run the following command to start the worker. The worker is configured to connect to the proxy.
```
go run ./worker
```
5) Run the following command to start the example. The client the starter uses is configured to connect to the proxy.
```
go run ./starter
```
6) Run the following command and see that when Temporal CLI is connected directly to Temporal it cannot display the payloads as they are encoded (compressed)
```
temporal workflow show --workflow-id grpcproxy_workflowID
```
7) Run the following command to see that when Temporal CLI is connected to Temporal via the proxy it can display the payloads
```
temporal workflow show --workflow-id grpcproxy_workflowID --address 'localhost:8081'
```
