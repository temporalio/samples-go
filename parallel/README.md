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

The workflow will:
- start and wait for two signals named "branch1" and "branch2"
- print to the screen the workflow-id and run-id (this information will be required for next step)
```
2022/10/21 12:54:58 Started workflow ca1b9934-9cbf-4b2c-9339-f19b29147ef6 426c1224-a537-40a6-8c56-61b38aba144c
```

4) copy the previous step's workflow-id and run-id signal the workflow to complete "branch1" or "branch2"

```shell script
# to complete branch 1
go run parallel/signaler/main.go <workflow-id> <run-id> branch1
# e.g. go run parallel/signaler/main.go ca1b9934-9cbf-4b2c-9339-f19b29147ef6 426c1224-a537-40a6-8c56-61b38aba144c branch1

# to complete branch 2
go run parallel/signaler/main.go <workflow-id> <run-id> branch2
# e.g. go run parallel/signaler/main.go ca1b9934-9cbf-4b2c-9339-f19b29147ef6 426c1224-a537-40a6-8c56-61b38aba144c branch2
```