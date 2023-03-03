## Infrequent polling

This sample shows how we can use Activity retries for infrequent polling of a third-party service (for example via REST).
This method can be used for infrequent polls of one minute or slower.

We utilize activity retries for this option, setting Retries options:
* setBackoffCoefficient to 1
* setInitialInterval to the polling interval (in this sample set to 60 seconds)

This will allow us to retry our Activity exactly on the set interval.

Since our test service simulates it being "down" for 4 polling attempts and then returns "OK" on the 5th poll attempt, our Workflow is going to perform 4 activity retries with a 60 second poll interval, and then return the service result on the successful 5th attempt.

Note that individual Activity retries are not recorded in Workflow History, so we this approach we can poll for a very long time without affecting the history size.


### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker
```
go run infrequent/worker/main.go
```
3) Run the following command to start the example
```
go run infrequent/starter/main.go
```
