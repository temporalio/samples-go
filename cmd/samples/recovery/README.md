### Recovery Sample
This sample implements a RecoveryWorkflow which is designed to restart all TripWorkflow executions which are currently
outstanding and replay all signals from previous run.  This is useful where a bad code change is rolled out which
causes workflows to get stuck or state is corrupted.

### Steps to run this sample
1) Run the following command to start worker
```
./bin/query -m worker
```
2) Run the following command to start trip workflow
```
./bin/recovery -m trigger -w UserA -wt main.TripWorkflow
```
3) Run the following command to query trip workflow
```
./bin/recovery -m query -w UserA
```
4) Run the following command to send signal to trip workflow
```
./bin/recovery -m signal -w UserA -s '{"ID": "Trip1", "Total": 10}'
```
4) Run the following command to start recovery workflow
```
./bin/recovery -m trigger -w UserB -wt recoveryworkflow -i '{"Type": "TripWorkflow", "Concurrency": 2}'
```