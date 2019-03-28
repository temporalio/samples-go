This sample workflow demos how to use query API to get the current state of running workflow.

query_workflow.go shows how to setup a custom workflow query handler
query_workflow_test.go shows how to unit-test query functionality

Steps to run this sample:
1) You need a cadence service running. See details in cmd/samples/README.md
2) Run the following command to start worker
```
./bin/query -m worker
```
3) Run the following command to trigger a workflow execution. You should see workflowID and runID print out on screen.
```
./bin/query -m trigger
```
4) Run "./bin/query -m query -w my_workflow_id -r my_run_id -t state" replace my_workflow_id and my_run_id with the workflowID and runID that you see in step 3. You should see current workflow state print on screen.
```
./bin/query -m query -w <workflow_id from step 3> -r <run_id from step 3> -t state
```
5) You could also replace the query type "state" to "__stack_trace" (replace -t state to -t __stack_trace) to dump the call stack for the workflow.
```
./bin/query -m query -w <workflow_id from step 3> -r <run_id from step 3> -t __stack_trace
```
