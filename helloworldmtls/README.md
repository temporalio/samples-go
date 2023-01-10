### Steps to run this sample:
1) Configure a [Temporal Server](https://github.com/temporalio/samples-go/tree/main/#how-to-use) (such as Temporal Cloud) with mTLS.

2) Run the following command to start the worker
```
go run ./helloworldmtls/worker -target-host my.namespace.tmprl.cloud:7233 -namespace my.namespace -client-cert path/to/cert.pem -client-key path/to/key.pem
```
3) Run the following command to start the example
```
go run ./helloworldmtls/starter -target-host my.namespace.tmprl.cloud:7233 -namespace my.namespace -client-cert path/to/cert.pem -client-key path/to/key.pem
```

If the server uses self-signed certificates and does not have the SAN set to the actual host, pass one of the following two options when starting the worker or the example above:
1. `-server-name` and provide the common name contained in the self-signed server certificate
2. `-insecure-skip-verify` which disables certificate and host name validation
