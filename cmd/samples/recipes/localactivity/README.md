This sample workflow demos how to use local activity to execute short/quick operations efficiently.

local_activity_workflow.go shows how to use local activity
local_activity_workflow_test.go shows how to unit-test workflow with local activity

Steps to run this sample:
1) You need a cadence service running. See details in cmd/samples/README.md
2) Run the following command to start worker
```
./bin/localactivity -m worker
```
3) Run the following command to trigger a workflow execution. You should see workflowID and runID print out on screen.
```
./bin/localactivity -m trigger
```
4) Run the following command to send signal "_1_" to the running workflow. You should see output that indicate 5 local activity has been run to check the conditions and one condition will be true which result in one activity to be scheduled.
```
./bin/localactivity -m signal -s _1_ -w <workflow ID from step 3>
```
5) Repeat step 4, but with different signal data, for example, send signal like _2_4_ to make 2 conditions true.
```
./bin/localactivity -m signal -s _2_4_ -w <workflow ID from step 3>
```
6) Run the following command this will exit the workflow.
```
./bin/localactivity -m signal -s exit
```