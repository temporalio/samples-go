### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run choice/worker/main.go
```
3) Run the following command to start the exclusive choice workflow
```
go run choice/exclusive_choice_starter/main.go
```
4) Run the following command to start the multi choice workflow
```
go run choice/multi_choice_starter/main.go -c multi
```
