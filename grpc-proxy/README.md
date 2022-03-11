### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run worker/main.go
```
3) Run the following command to start the example
```
go run starter/main.go
```
4) Run the following command and see that tctl cannot display the payloads as they are encoded (compressed)
```
tctl workflow show --wid grpcproxy_workflowID
```
5) Run the following command to start the remote codec server
```
go run proxy-server/main.go
```
6) Run the following command to see that tctl can now decode (uncompress) the payloads via the remote codec server
```
tctl --address 'localhost:8081' workflow show --wid grpcproxy_workflowID
```
