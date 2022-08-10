### Steps to run this sample:
1) You need a Temporal server configured with mTLS such as Temporal cloud.
2) Run the following command to start the worker
```
go run ./helloworldmtls/worker -target-host my.namespace.tmprl.cloud:7233 -namespace my.namespace -client-cert path/to/cert.pem -client-key path/to/key.pem
```
3) Run the following command to start the example
```
go run ./helloworldmtls/starter -target-host my.namespace.tmprl.cloud:7233 -namespace my.namespace -client-cert path/to/cert.pem -client-key path/to/key.pem
```
