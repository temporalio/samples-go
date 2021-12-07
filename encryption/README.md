### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Compile the encryption plugin for tctl
```
go build -o ../bin/encryption-plugin plugin/main.go
```
3) Run the following command to start the worker
```
go run worker/main.go
```
4) Run the following command to start the example
```
go run starter/main.go
```
5) Run the following command and see the encrypted payloads
```
export PATH="../bin:$PATH" TEMPORAL_CLI_PLUGIN_DATA_CONVERTER=encryption-plugin
tctl workflow show --wid encryption_workflowID
```
Note: plugins should normally be available in your PATH, we include the current directory in the path here for ease of testing.
