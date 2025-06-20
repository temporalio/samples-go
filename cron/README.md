This sample demonstrates how to setup cron based workflow.


**We recommend using [Schedules](../schedule) instead of Cron Jobs. Schedules were built to provide a better developer experience, including more configuration options and the ability to update or pause running Schedules.**

Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run 
```
go run cron/worker/main.go 
```
to start worker for cron workflow.
3) Run
```
go run cron/starter/main.go
```
to start workflow with cron expression scheduled to run every minute.
