### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```shell script
go run parallel/worker/main.go
```
3) Run the following command to start parallel workflow
```shell script
go run parallel/starter/main.go
```
