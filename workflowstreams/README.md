### Workflow Streams

A **workflow stream** is a durable publish/subscribe log hosted inside a Temporal
workflow, provided by the [`workflowstreams`](https://pkg.go.dev/go.temporal.io/sdk/contrib/workflowstreams)
contrib package. External code (activities, starters, other workflows) publishes
messages to named topics via **signals**; subscribers long-poll for new items via
**updates**; a **query** exposes the current offset. Because it is backed by
Temporal's durable execution, delivery is ordered, durable, and exactly-once, with
client-side batching, publisher dedup, continue-as-new survival, and truncation.

### Key APIs

Workflow side — construct a stream once at the start of the workflow and publish to topics:

```go
stream, err := workflowstreams.NewWorkflowStream(ctx, input.StreamState)
_ = stream.Topic("status").Publish(StatusEvent{Kind: "received"})
```

Client side (activities, starters, external code) — publish and subscribe:

```go
c := workflowstreams.NewClient(temporalClient, workflowID, workflowstreams.Options{})
defer c.Close(ctx)

for item, err := range c.Subscribe(ctx, workflowstreams.SubscribeOptions{Topics: []string{"status"}}) {
    if err != nil { /* handle */ }
    var evt StatusEvent
    _ = converter.GetDefaultDataConverter().FromPayload(item.Data, &evt)
}
```

Offsets are **global** across topics. To resume a subscription from where a
previous one left off, set `FromOffset` in `SubscribeOptions` to one past the
last item you consumed:

```go
for item, err := range c.Subscribe(ctx, workflowstreams.SubscribeOptions{
    Topics:     []string{"status"},
    FromOffset: lastItem.Offset + 1, // zero (the default) starts from the beginning
}) {
    // ...
}
```

### Steps to run this sample

1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use)
   (for example, `temporal server start-dev`).

2) Start the worker (serves scenarios 1–4):

```
go run ./workflowstreams/worker
```

3) Run any of the scenarios below in a separate terminal.

#### Scenario 1 — basic publish/subscribe

An order workflow publishes status events itself while an activity publishes
fine-grained progress events to the same stream. A subscriber consumes both topics.

```
go run ./workflowstreams/publisher
```

Expected output (interleaving may vary):

```
[status]   received: order=order-42
[progress] charging card...
[progress] card charged
[status]   shipped: order=order-42
[progress] charge id: charge-order-42
[status]   complete: order=order-42
```

#### Scenario 2 — reconnecting subscriber

A subscriber reads a few pipeline stage events, disconnects, then a brand-new
client resumes from the saved offset without missing events or seeing duplicates.

```
go run ./workflowstreams/reconnecting
```

#### Scenario 3 — external publisher

The hub workflow does no work of its own; it just hosts the stream. A separate
publisher pushes news into it (using the same client factory used to subscribe) and
then signals the workflow to close. Here a publisher and subscriber run concurrently.

```
go run ./workflowstreams/external-publisher
```

#### Scenario 4 — truncating ticker

The ticker workflow periodically truncates old entries to bound its history, trading
complete history for a bounded log. A *fast* subscriber that reads from the start keeps
up and sees every tick. A *late* subscriber joins after truncation and resumes from a
stale offset; the stream fast-forwards it to the current base offset, so it cannot see
the truncated ticks.

```
go run ./workflowstreams/truncating-ticker
```

Expected output (the late subscriber's first line shows the fast-forward):

```
[late] requested offset 1 but it was truncated; fast-forwarded to offset 5 (skipped 4 tick(s))
[late] offset=  5  n=5
...
```

#### Scenario 5 — LLM token streaming

The workflow hosts the stream while an activity makes a streaming OpenAI call and
republishes each token delta. On a retry it emits a retry event and the subscriber
rewinds the terminal and re-renders. This scenario runs on its own worker and task
queue, and requires `OPENAI_API_KEY`.

```
# Terminal A
OPENAI_API_KEY=sk-... go run ./workflowstreams/llmworker

# Terminal B
go run ./workflowstreams/llm "Explain durable execution in one sentence."
```
