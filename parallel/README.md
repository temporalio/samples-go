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

the workflow will start and wait for two signals named "branch1" and "branch2"

4) get the previous step's workflow-id and run-id signal the workflow to complete "branch1" or "branch2"

```shell script
# to complete branch 1
go run parallel/signaler/main.go <workflow-id> <run-id> branch1

# to complete branch 2
go run parallel/signaler/main.go <workflow-id> <run-id> branch2
```