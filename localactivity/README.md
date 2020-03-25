This sample workflow demos how to use local activity to execute short/quick operations efficiently.

local_activity_workflow.go shows how to use local activity
local_activity_workflow_test.go shows how to unit-test workflow with local activity

Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start worker
```
go run localactivity/worker/main.go
```
3) Run the following command to trigger a workflow execution. You should see workflowID and runID print out on screen.
```
go run localactivity/starter/main.go
```
4) Run the following command to send signal `_1_` to the running workflow. You should see output that indicate 5 local activity has been run to check the conditions and one condition will be true which result in one activity to be scheduled.
```
go run localactivity/signal/main.go -s _1_
```
5) Repeat step 4, but with different signal data, for example, send signal like `_2_4_` to make 2 conditions true.
```
go run localactivity/signal/main.go -s _2_4_
```
6) Run the following command this will exit the workflow.
```
go run localactivity/signal/main.go -s exit
```