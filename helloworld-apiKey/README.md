### Steps to run this sample:
1) Configure a [Temporal Server](https://github.com/temporalio/samples-go/tree/main/#how-to-use) (such as Temporal Cloud) with apiKey.

2) Run the following command to start the worker
```
go run ./helloworld-apiKey/worker \
    -target-host my.namespace.tmprl.cloud:7233 \
    -namespace my.namespace \
    -api-key CLIENT_API_KEY
```
3) Run the following command to start the example
```
go run ./helloworld-apiKey/starter \
    -target-host my.namespace.tmprl.cloud:7233 \
    -namespace my.namespace \
    -api-key CLIENT_API_KEY
```