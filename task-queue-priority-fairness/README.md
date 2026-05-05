# Task Queue Priority and Fairness

This sample demonstrates Temporal Task Queue Priority and Fairness for Activity Tasks using a multi-tenant media rendering workload.

A SaaS rendering platform receives urgent previews, normal renders, and background archive jobs from multiple tenants into one shared Activity Task Queue. The sample intentionally creates backlog so dispatch order is visible:

- `PriorityKey` chooses which priority level is dispatched first. Lower values run first, so urgent preview jobs use `PriorityKey=1`, normal render jobs use `PriorityKey=3`, and background archive jobs use `PriorityKey=5`.
- `FairnessKey` groups work by tenant within a priority level. Every Activity Task in this sample uses its tenant ID as a non-empty fairness key.
- `FairnessWeight` gives `premium-media` a larger proportional share while it has backlog. The premium tenant uses `FairnessWeight=3.0`; regular tenants use `FairnessWeight=1.0`.

Priority determines which priority sub-queue Tasks go into. Fairness determines ordering within a given priority level. Fairness ordering is probabilistic and observational, so the exact order can vary between runs.

## Run the backlog demo

### 1. Start Temporal Server with Fairness enabled

The following dev server config is known to work well for this sample:

```bash
temporal server start-dev \
  --dynamic-config-value matching.useNewMatcher=true \
  --dynamic-config-value matching.enableFairness=true \
  --dynamic-config-value matching.numTaskqueueReadPartitions=1 \
  --dynamic-config-value matching.numTaskqueueWritePartitions=1
```

### 2. Start only the Workflow Worker

```bash
go run task-queue-priority-fairness/worker/main.go -mode workflow
```

### 3. Start the Workflow

In another terminal:

```bash
go run task-queue-priority-fairness/starter/main.go
```

The starter prints a reminder to start the Activity Worker. At this point, the Workflow has scheduled many Activity Tasks, but no Activity Worker is polling yet, so the Activity Task Queue has backlog.

### 4. Start the constrained Activity Worker

In another terminal:

```bash
go run task-queue-priority-fairness/worker/main.go -mode activity
```

The Activity Worker uses `MaxConcurrentActivityExecutionSize: 1`, making the dispatch/start order easy to observe in logs and in the starter output.

## What to look for

The starter prints a table of observed Activity start order and a summary. In the backlog-focused flow, urgent preview jobs submitted last should be dispatched before lower-priority queued work. Within normal render jobs, small tenants should appear before the large tenant drains its backlog, and `premium-media` should receive repeated dispatches while it remains backlogged.

## Example output

A successful run should look similar to this. 
> [!NOTE]
> Timestamps, Run IDs, and Worker IDs will differ on your machine.

```text
2026/05/05 00:45:09 Started workflow. WorkflowID task-queue-priority-fairness-<example-id> RunID <example-run-id>
If you are running the full backlog demo, start the Activity Worker now:
go run task-queue-priority-fairness/worker/main.go -mode activity
Activity start order:

01 started_at=<ts> priority=1 fairness_key=premium-media  weight=3.0 kind=urgent-preview      job=premium-media-urgent-preview-00
02 started_at=<ts> priority=1 fairness_key=premium-media  weight=3.0 kind=urgent-preview      job=premium-media-urgent-preview-01
03 started_at=<ts> priority=1 fairness_key=small-studio-a weight=1.0 kind=urgent-preview      job=small-studio-a-urgent-preview-01
04 started_at=<ts> priority=1 fairness_key=small-studio-a weight=1.0 kind=urgent-preview      job=small-studio-a-urgent-preview-00
05 started_at=<ts> priority=3 fairness_key=premium-media  weight=3.0 kind=normal-render       job=premium-media-normal-render-06
06 started_at=<ts> priority=3 fairness_key=small-studio-b weight=1.0 kind=normal-render       job=small-studio-b-normal-render-00
07 started_at=<ts> priority=3 fairness_key=small-studio-a weight=1.0 kind=normal-render       job=small-studio-a-normal-render-02
...
37 started_at=<ts> priority=3 fairness_key=large-studio   weight=1.0 kind=normal-render       job=large-studio-normal-render-17
38 started_at=<ts> priority=5 fairness_key=large-studio   weight=1.0 kind=background-archive  job=large-studio-background-archive-03
39 started_at=<ts> priority=5 fairness_key=large-studio   weight=1.0 kind=background-archive  job=large-studio-background-archive-02
40 started_at=<ts> priority=5 fairness_key=large-studio   weight=1.0 kind=background-archive  job=large-studio-background-archive-01
41 started_at=<ts> priority=5 fairness_key=large-studio   weight=1.0 kind=background-archive  job=large-studio-background-archive-00

Summary:

Priority:
  Urgent jobs use PriorityKey=1.
  Normal jobs use PriorityKey=3.
  Background jobs use PriorityKey=5.
  Urgent work was dispatched before lower-priority queued work: OBSERVED

Fairness:
  Tenant IDs are used as FairnessKey values.
  Small tenants appeared before the large tenant drained all normal jobs: OBSERVED

Weighted fairness:
  premium-media uses FairnessWeight=3.0.
  Other tenants use FairnessWeight=1.0.
  Premium work received repeated dispatches while backlogged: OBSERVED

Note:
  Fairness ordering is probabilistic and can vary between runs.
```

How to read this output:

- Rows `01` to `04` are all urgent (`priority=1`) even though urgent jobs were scheduled last in `BuildJobs()`. This shows priority overtaking queued lower-priority work.
- Normal jobs (`priority=3`) include multiple fairness keys near the front (`premium-media`, `small-studio-a`, `small-studio-b`, `large-studio`) instead of draining one tenant first.
- `premium-media` appears more often in the early normal-priority dispatches while it has queued work, reflecting its `FairnessWeight=3.0` compared with `1.0` for the other tenants.
- Background jobs (`priority=5`) start only after normal-priority work in this run, matching the expected priority behavior.
