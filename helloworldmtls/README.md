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

If the server uses self-signed certificates, pass one of the following two options when starting the worker or the example above:
1. `-server-name` and provide the common name contained in the self-signed server certificate
2. `-insecure-skip-verify` which disables certificate and host name validation
