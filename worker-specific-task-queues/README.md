# Sticky Activity Queues

This sample shows how to have [Sticky Execution](https://docs.temporal.io/tasks/#sticky-execution): using a unique task queue per Worker to have certain activities only run on that specific Worker.

The strategy is:

- Create a `StickyTaskQueue.GetStickyTaskQueue` activity that generates a unique task queue name, `uniqueWorkerTaskQueue`.
- It doesn't matter where this activity is run, so it can be "non sticky" as per Temporal default behavior.
- In this demo, `uniqueWorkerTaskQueue` is simply a `uuid` initialized in the Worker, but you can inject smart logic here to uniquely identify the Worker, [as Netflix did](https://community.temporal.io/t/using-dynamic-task-queues-for-traffic-routing/3045).
- For activities intended to be "sticky", only register them in one Worker, and have that be the only Worker listening on that `uniqueWorkerTaskQueue`.
- Execute workflows from the Client like normal.

Activities have been artificially slowed with `time.Sleep(3 * time.Second)` to simulate slow activities.

### Running this sample

```bash
go run activities-sticky-queues/worker/main.go
```

Start the Workflow Execution:

```bash
go run activities-sticky-queues/starter/main.go
```
