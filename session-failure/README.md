This sample workflow demos how to recover from a session failure inside a workflow

The workflow first creates a session then starts a short activity meant to simulate preparing the session session, then it starts a long running activity on the session worker. If the session worker goes down for any reason the session will fail to heartbeat and be marked as failed. This will cause any activities running on the session to be cancelled and the workflow to retry the whole sequence on a new session after a timeout.

### Note on session failure: 

Workflows detect a session worker has gone down through heartbeats by the session worker, so the workflow has a stale view of the session workers state. This is important to consider if your
workflow schedules any activities on a session that can fail due to a timeout. It is possible that when a session worker fails, if your activities timeout is shorter than twice the session heartbeat timeout, your activity may fail with a timeout error and the session state will not be failed yet.

It is also worth noting if a session worker is restarted then it is considered a new session worker and will not pick up any activities scheduled on the old session worker. If you want to be able to keep scheduling activities on the same host after restart look at ../activities-sticky-queues 

Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command multiple times on different console window. This is to simulate running workers on multiple different machines.
```
go run session-failure/worker/main.go
```
1) Run the following command to submit a start request for this session failure workflow.
```
go run session-failure/starter/main.go
```
1) If you want to observe the workflow recover from a failed session you can restart 
the worker you launched in step 2).

You should see that all activities for one particular workflow execution are scheduled to run on one console window.
