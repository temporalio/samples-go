This sample shows how to connect a client to Temporal using mtls where the certificates are dynamically loaded. This allows the credentials to be replaced without restarting the worker. 

### Steps to run this sample:
1) Configure a [Temporal Server](https://github.com/temporalio/samples-go/tree/main/#how-to-use) (such as Temporal Cloud) with mTLS.

2) Run the following command to start the worker
```
go run ./dynamicmtls/worker -target-host my.namespace.tmprl.cloud:7233 -namespace my.namespace -client-cert path/to/cert.pem -client-key path/to/key.pem
```
3) Run the following command to start the example
```
go run ./dynamicmtls/starter -target-host my.namespace.tmprl.cloud:7233 -namespace my.namespace -client-cert path/to/cert.pem -client-key path/to/key.pem
```

Note:

If the server uses self-signed certificates and does not have the SAN set to the actual host, pass one of the following two options when starting the worker or the example above:
1. `-server-name` and provide the common name contained in the self-signed server certificate
2. `-insecure-skip-verify` which disables certificate and host name validation
