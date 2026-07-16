# Worker-Specific Task Queues

*[Sessions](https://docs.temporal.io/dev-guide/go/features#worker-sessions) are an alternative to Worker-specific Tasks Queues.*

Use a unique Task Queue for each Worker in order to have certain Activities run on a specific Worker.

This is useful in scenarios where multiple Activities need to run in the same process or on the same host, for example to share memory or disk. This sample has a file processing Workflow, where one Activity downloads the file to disk and other Activities process it and clean it up.

The strategy is:

- Each Worker process creates two `worker` instances:
  - One instance listens on the `shared-task-queue` Task Queue.
  - Another instance listens on a uniquely generated Task Queue (in this case, `uuid` is used, but you can inject smart logic here to uniquely identify the Worker, [as Netflix did](https://community.temporal.io/t/using-dynamic-task-queues-for-traffic-routing/3045)).
- The Workflow and the first Activity are run on `shared-task-queue`.
- The first Activity returns one of the uniquely generated Task Queues (that only one Worker is listening on—i.e. the **Worker-specific Task Queue**).
- The rest of the Activities do the file processing and are run on the Worker-specific Task Queue.

Activities have been artificially slowed with `time.Sleep(3 * time.Second)` to simulate slow activities.

### Running this sample

```bash
go run worker-specific-task-queues/worker/main.go
```

Start the Workflow Execution:

```bash
go run worker-specific-task-queues/starter/main.go
```

### Things to try
You can try to intentionally crash Workers while they are doing work to see what happens when work gets "stuck" in a unique queue: currently the Workflow will `scheduleToCloseTimeout` without a Worker, and retry when a Worker comes back online.

After the 5th attempt, it logs `Workflow failed after multiple retries.` and exits. But you may wish to implement compensatory logic, including notifying you.