## Frequent polling

This sample shows how we can implement frequent polling (1 second or faster) inside our Activity.
The implementation is a loop that polls our service and then sleeps for the poll interval (1 second in the sample).

To ensure that polling activity is restarted in a timely manner, we make sure that it heartbeats on every iteration.
Note that heartbeating only works if we set the HeartBeatTimeout to a shorter value than the activity
StartToClose timeout.

### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run frequent/worker/main.go
```
3) Run the following command to start the example
```
go run frequent/starter/main.go
```
