# Updatable Timer Sample

A helper structure that supports blocking sleep that can be rescheduled at any moment.

Demonstrates:

* Timer and its cancellation
* Signal Channel
* Selector used to wait on both timer and channel

### Steps to run this sample:

1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker

```
go run updatabletimer/worker/main.go
```

3) Run the following command to start the example

```
go run updatabletimer/starter/main.go
```

4)  Run the following command to update the timer wake-up time

```
go run updatabletimer/updater/main.go
```