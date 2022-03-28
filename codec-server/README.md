### Steps to run this sample:

This sample shows how to decode payloads that have been encoded by a codec so they can be displayed by tctl and Temporal Web.
The codec server supports OIDC authentication (via JWT in the Authorization header).
If you are using OIDC for Temporal Web this token can be passed on to the codec server, see:
https://github.com/temporalio/web/pull/445/files#diff-2eea834a27d42e5223553feb6a5795a37d859e2845df9a5b6b938a8f0a8271c4R23
Configuring OIDC is outside of the scope of this sample, but please see ../serverjwtauth for more details about authentication.

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
tctl workflow show --wid codecserver_workflowID
```
5) Run the following command to start the remote codec server
```
go run ./codec-server
```
6) Run the following command to see that tctl can now decode (uncompress) the payloads via the remote codec server
```
tctl --codec_endpoint 'http://localhost:8081/{namespace}' workflow show --wid codecserver_workflowID
```
