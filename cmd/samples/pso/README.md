This sample workflow demos a long iterative math optimization process using particle swarm optimization (PSO). 

The workflow first does some data structure initialization and then runs many iterations using a child workflow. The child workflow runs 10 iterations and then uses ContinueAsNew to avoid to store too long history in the Cadence database. In case of recovery the whole history has to be replayed to reconstruct the workflow state. So if history is too large the recover can take very long time.
Each particle is processed in parallel using worflow.Go and the math grunt work is done in the activites.
Since the data structure that maintains the optimization state has to be passed to the child workflow and the activities, a custom DataConverter has been implemented to take care of serialization/deserialization.
Also the query API is supported to get the current state of running workflow.

Steps to run this sample: 
1) You need a cadence service running. See details in cmd/samples/README.md
2) Run the following command multiple times on different console window. This is to simulate running workers on multiple different machines.
```
./bin/pso -m worker
```
3) Run the following command to submit a start request for this PSO workflow.
```
./bin/pso -m trigger
```
4) Query the state with
```
./bin/pso -m query -w <workflow_id from step 3> -r <run_id from step 3> -t state
```
Replace -t state with -t \_\_stack_trace to dump the call stack for the workflow.

You should see that all activities for one particular workflow execution are scheduled to run on one console window.
