### Sample overview:

This sample shows how to compress payloads using a GRPC proxy. This allows managing the codec(s) centrally rather than needing to configure each client/worker.

### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the GRPC proxy listening on port 8081
```
go run proxy-server/main.go
```
3) Run the following command to start the worker. The worker is configured to connect to the proxy.
```
go run worker/main.go
```
4) Run the following command to start the example. The client the starter uses is configured to connect to the proxy.
```
go run starter/main.go
```
5) Run the following command and see that when tctl is connected directly to Temporal it cannot display the payloads as they are encoded (compressed)
```
tctl workflow show --wid grpcproxy_workflowID
```
6) Run the following command to see that when tctl is connected to Temporal via the proxy it can display the payloads
```
tctl --address 'localhost:8081' workflow show --wid grpcproxy_workflowID
```
