This sample workflow demos a long iterative math optimization process using particle swarm optimization (PSO). 

The workflow first does some data structure initialization and then runs many iterations using a child workflow. The child workflow runs 10 iterations and then uses `ContinueAsNew` to avoid to store too long history in the Temporal database. In case of recovery the whole history has to be replayed to reconstruct the workflow state. So if history is too large the recover can take very long time.
Each particle is processed in parallel using `worflow.Go` and the math grunt work is done in the activites.
Since the data structure that maintains the optimization state has to be passed to the child workflow and the activities, a custom `DataConverter` has been implemented to take care of serialization/deserialization.
Also the query API is supported to get the current state of running workflow.

Steps to run this sample: 
1) You need a Temporal service running. See details in README.md
2) Run the following command multiple times on different console window. This is to simulate running workers on multiple different machines.
```
go run pso/worker/main.go
```
3) Run the following command to submit a start request for this PSO workflow.
```
go run pso/starter/main.go
```
4) Query the call stack for the workflow with
```
go run pso/query/main.go -w <workflow_id from step 3> -r <run_id from step 3>
```
