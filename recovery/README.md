
### Recovery Sample
This sample implements a `RecoveryWorkflow` which is designed to restart all `TripWorkflow` executions which are currently
outstanding and replay all signals from previous run.  This is useful where a bad code change is rolled out which
causes workflows to get stuck or state is corrupted.

### Steps to run this sample
1) Run the following command to start worker
```
go run recovery/worker/main.go
```
2) Run the following command to start trip workflow
```
go run recovery/starter/main.go
```
3) Run the following command to query trip workflow
```
go run recovery/query/main.go
```
4) Run the following command to send signal to trip workflow
```
go run recovery/signal/main.go -s '{"ID": "Trip1", "Total": 10}'
```
4) Run the following command to start recovery workflow
```
go run recovery/starter/main.go -w recovery_workflow -wt recoveryworkflow -i '{"Type": "TripWorkflow", "Concurrency": 2}'
```