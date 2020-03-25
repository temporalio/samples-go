### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run choice/worker/main.go
```
3) Run the following command to start the single choice workflow
```
go run choice/starter/main.go
```
4) Run the following command to start the multi choice workflow
```
go run choice/starter/main.go -c multi
```
