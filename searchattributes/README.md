### Steps to run this sample:
1) Before running this, Temporal Server need to run with advanced visibility store. 
See https://github.com/temporalio/temporal/blob/master/docs/visibility-on-elasticsearch.md
2) Run the following command to start the worker
```
go run searchattributes/worker/main.go
```
3) Run the following command to start the example
```
go run searchattributes/starter/main.go
```
