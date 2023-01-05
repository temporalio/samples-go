# Temporal Go SDK samples

[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B18405%2Fgithub.com%2Ftemporalio%2Fsamples-go.svg?type=shield)](https://app.fossa.com/projects/custom%2B18405%2Fgithub.com%2Ftemporalio%2Fsamples-go?ref=badge_shield)

This repository contains several sample Workflow applications that demonstrate the various capabilities of the Temporal
Server via the Temporal Go SDK.

- Temporal Server repo: [https://github.com/temporalio/temporal](https://github.com/temporalio/temporal)
- Temporal Go SDK repo: [https://github.com/temporalio/sdk-go](https://github.com/temporalio/sdk-go)
- Go SDK docs: https://docs.temporal.io/docs/go/introduction

## How to use

- Run this in the browser with
  Gitpod: [![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-908a85?logo=gitpod)](https://gitpod.io/#https://github.com/temporalio/samples-go/)
- Or run Temporal Server locally with [VSCode Remote Containers](https://code.visualstudio.com/docs/remote/containers)
  . [![Open in Remote - Containers](https://img.shields.io/static/v1?label=Remote%20-%20Containers&message=Open&color=blue&logo=visualstudiocode)](https://vscode.dev/redirect?url=vscode://ms-vscode-remote.remote-containers/cloneInVolume?url=https://github.com/temporalio/samples-go)
- Lastly, you can run Temporal Server locally on your own (follow
  the [Quick install guide](https://docs.temporal.io/clusters/quick-install)), then clone this repository

The [helloworld](https://github.com/temporalio/samples-go/tree/main/helloworld) sample is a good place to start.

## Samples directory

Each sample demonstrates one feature of the SDK, together with tests.

- [**Basic hello world**](https://github.com/temporalio/samples-go/tree/main/helloworld): Simple example of a Workflow
  Definition and an Activity Definition.

- [**Basic mTLS hello world**](https://github.com/temporalio/samples-go/tree/main/helloworldmtls): Simple example of a
  Workflow Definition and an Activity Definition using mTLS like Temporal Cloud.

### API demonstrations

- **Async activity completion**: Example of
  an [Expense reporting](https://github.com/temporalio/samples-go/tree/main/expense) Workflow that communicates with a
  server API. Additional
  documentation: [How to complete an Activity Execution asynchronously in Go](https://docs.temporal.io/application-development/foundations/#develop-activities)

- [**Retry Activity Execution**](https://github.com/temporalio/samples-go/tree/main/retryactivity): This samples
  executes an unreliable Activity. The Activity is executed with a custom Retry Policy. If the Activity Execution fails,
  the Server will schedule a retry based on the Retry Policy. This Activity also includes a Heartbeat, which enables it
  to resume from the Activity Execution's last reported progress when it retries.

- [**Child Workflow**](https://github.com/temporalio/samples-go/tree/main/child-workflow): Demonstrates how to use
  execute a Child Workflow from a Parent Workflow Execution. A Child Workflow Execution only returns to the Parent
  Workflow Execution after completing its last Run.

- [**Child Workflow with
  ContinueAsNew**](https://github.com/temporalio/samples-go/tree/main/child-workflow-continue-as-new): Demonstrates
  that the call to Continue-As-New, by a Child Workflow Execution, is *not visible to the a parent*. The Parent Workflow
  Execution receives a notification only when a Child Workflow Execution completes, fails or times out. This is a useful
  feature when there is a need to **process a large set of data**. The child can iterate over the data set calling
  Continue-As-New periodically without polluting the parents' history.

- [**Cancellation**](https://github.com/temporalio/samples-go/tree/main/cancellation): Demonstrates how to cancel a
  Workflow Execution by calling `CancelWorkflow`, an how to defer an Activity Execution that "cleans up" after the
  Workflow Execution has been cancelled.

- **Coroutines**: Do not use native `go` routines in Workflows. Instead use Temporal coroutines (`workflow.Go()`) to
  maintain a [deterministic](https://docs.temporal.io/application-development/foundations/#develop-workflows) Workflow. Can be
  seen in the [Goroutine](https://github.com/temporalio/samples-go/tree/main/goroutine)
  , [DSL](https://github.com/temporalio/samples-go/tree/main/dsl)
  , [Recovery](https://github.com/temporalio/samples-go/tree/main/recovery)
  , [PSO](https://github.com/temporalio/samples-go/tree/main/pso) Workflow examples.

- [**Cron Workflow**](https://github.com/temporalio/samples-go/tree/main/cron): Demonstrates a recurring Workflow
  Execution that occurs according to a cron schedule. This samples showcases the `HasLastCompletionResult`
  and `GetLastCompletionResult` APIs which are used to pass information between executions. Additional
  documentation: [What is a Temporal Cron Job?](https://docs.temporal.io/docs/content/what-is-a-temporal-cron-job).

- [**Encryption**](https://github.com/temporalio/samples-go/tree/main/encryption): How to use encryption for
  Workflow/Activity data with the DataConverter API. Also includes an example of stacking encoders (in this case
  encryption and compression)

- [**Codec Server**](https://github.com/temporalio/samples-go/tree/main/codec-server): Demonstrates using a codec
  server to decode payloads for display in tctl and Temporal Web. This setup can be used for any kind of codec, common
  examples are compression or encryption.

- [**Query Example**](https://github.com/temporalio/samples-go/tree/main/query): Demonstrates how to Query the state
  of a single Workflow Execution using the `QueryWorkflow` and `SetQueryHandler` APIs. Additional
  documentation: [How to Query a Workflow Execution in Go](https://docs.temporal.io/application-development/features/#queries).

- **Selectors**: Do not use the native Go `select` statement. Instead
  use [Go SDK Selectors](https://docs.temporal.io/docs/go/selectors) (`selector.Select(ctx)`) to maintain
  a [deterministic](https://docs.temporal.io/application-development/foundations/#develop-workflows) Workflow. Can be seen in
  the [Pick First](https://github.com/temporalio/samples-go/tree/main/pickfirst)
  , [Mutex](https://github.com/temporalio/samples-go/tree/main/mutex)
  , [DSL](https://github.com/temporalio/samples-go/tree/main/dsl),
  and [Timer](https://github.com/temporalio/samples-go/tree/main/timer) examples.

- **Sessions**: Demonstrates how to bind a set of Activity Executions to a specific Worker after the first Activity
  executes. This feature is showcased in
  the [File Processing example](https://github.com/temporalio/samples-go/tree/main/fileprocessing). Addition
  documentation: [How to use Sessions in Go](https://docs.temporal.io/go/how-to-create-a-worker-session-in-go).

- **Signals**: Can be seen in the [Recovery](https://github.com/temporalio/samples-go/tree/main/recovery)
  and [Mutex](https://github.com/temporalio/samples-go/tree/main/mutex) examples. Additional
  documentation: [eCommerce application tutorial](https://learn.temporal.io/tutorials/go/ecommerce/)
  , [How to send and handle Signals in Go](https://docs.temporal.io/application-development/features/#signals)
  .

- [**Memo**](https://github.com/temporalio/samples-go/tree/main/memo): Demonstrates how to use Memo that can be used
  to store any kind of data.

- [**Search Attributes**](https://github.com/temporalio/samples-go/tree/main/searchattributes): Demonstrates how to
  use custom Search Attributes that can be used to find Workflow Executions using predicates (must use
  with [Elasticsearch](https://docs.temporal.io/clusters/how-to-integrate-elasticsearch-into-a-temporal-cluster)).

- [**Timer Futures**](https://github.com/temporalio/samples-go/tree/main/timer): The sample starts a long running
  order processing operation and starts a Timer (`workflow.NewTimer()`). If the processing time is too long, a
  notification email is "sent" to the user regarding the delay (the execution does not cancel). If the operation
  finishes before the Timer fires, then the Timer is cancelled.

- [**Tracing and Context Propagation**](https://github.com/temporalio/samples-go/tree/main/ctxpropagation):
  Demonstrates the client initialization with a context propagator, which propagates specific information in
  the `context.Context` object across the Workflow Execution. The `context.Context` object is populated with information
  prior to calling `StartWorkflow`. This example demonstrates that the information is available in the Workflow
  Execution and Activity Executions. Additional
  documentation: [How to use tracing in Go](https://docs.temporal.io/go/tracing).

- [**Updatable Timer**](https://github.com/temporalio/samples-go/tree/main/updatabletimer): Demonstrates timer
  cancellation and use of a Selector to wait on a Future and a Channel simultaneously.

- [**Greetings**](https://github.com/temporalio/samples-go/tree/main/greetings): Demonstrates how to pass dependencies
  to activities defined as struct methods.

- [**Greetings Local**](https://github.com/temporalio/samples-go/tree/main/greetingslocal): Demonstrates how to pass
  dependencies to local activities defined as struct methods.

- [**Interceptors**](https://github.com/temporalio/samples-go/tree/main/interceptor): Demonstrates how to use
  interceptors to intercept calls, in this case for adding context to the logger.

### Dynamic Workflow logic examples

These samples demonstrate some common control flow patterns using Temporal's Go SDK API.

- [**Dynamic Execution**](https://github.com/temporalio/samples-go/tree/main/dynamic): Demonstrates how to execute
  Workflows and Activities using a name rather than a strongly typed function.

- [**Branching Acitivties**](https://github.com/temporalio/samples-go/tree/main/branch): Executes multiple Activities
  in parallel. The number of branches is controlled by a parameter that is passed in at the start of the Workflow
  Execution.

- [**Exclusive Choice**](https://github.com/temporalio/samples-go/tree/main/choice-exclusive): Demonstrates how to
  execute Activities based on a dynamic input.

- [**Multi-Choice**](https://github.com/temporalio/samples-go/tree/main/choice-multi): Demonstrates how to execute
  multiple Activities in parallel based on a dynamic input.

- [**Mutex Workflow**](https://github.com/temporalio/samples-go/tree/main/mutex): Demonstrates the ability to
  lock/unlock a particular resource within a particular Temporal Namespace. In this examples the other Workflow
  Executions within the same Namespace wait until a locked resource is unlocked. This shows how to avoid race conditions
  or parallel mutually exclusive operations on the same resource.

- [**Goroutine Workflow**](https://github.com/temporalio/samples-go/tree/main/goroutine): This sample executes
  multiple sequences of activities in parallel using the `workflow.Go()` API.

- [**Pick First**](https://github.com/temporalio/samples-go/tree/main/pickfirst): This sample executes Activities in
  parallel branches, picks the result of the branch that completes first, and then cancels other Activities that have
  not finished.

- [**Split/Merge Future**](https://github.com/temporalio/samples-go/tree/main/splitmerge-future): Demonstrates how to
  use futures to await for completion of multiple activities invoked in parallel. This samples to processes chunks of a
  large work item in parallel, and then merges the intermediate results to generate the final result.
-
- [**Split/Merge Selector**](https://github.com/temporalio/samples-go/tree/main/splitmerge-selector): Demonstrates how
  to use Selector to process activity results as soon as they become available. This samples to processes chunks of a
  large work item in parallel, and then merges the intermediate results to generate the final result.

- [**Synchronous Proxy Workflow pattern**](https://github.com/temporalio/samples-go/tree/main/synchronous-proxy): This
  sample demonstrates a synchronous interaction with a "main" Workflow Execution from a "proxy" Workflow Execution. The
  proxy Workflow Execution sends a Signal to the "main" Workflow Execution, then blocks, waiting for a Signal in
  response.

- [**Saga pattern**](https://github.com/temporalio/samples-go/tree/main/saga): This sample demonstrates how to implement
  a saga pattern using golang defer feature.

- [**Await for signal processing**](https://github.com/temporalio/samples-go/tree/main/await-signals): Demonstrates how
  to process out of order signals processing using `Await` and `AwaitWithTimeout`.

### Scenario based examples

- [**DSL Workflow**](https://github.com/temporalio/samples-go/tree/main/dsl): Demonstrates how to implement a
  DSL-based Workflow. This sample contains 2 yaml files that each define a custom "workflow" which instructs the
  Temporal Workflow. This is useful if you want to build in a "low code" layer.

- [**Expense Request**](https://github.com/temporalio/samples-go/tree/main/expense): This demonstrates how to process
  an expense request. This sample showcases how to complete an Activity Execution asynchronously.

- [**File Processing**](https://github.com/temporalio/samples-go/tree/main/fileprocessing): Demonstrates how to
  download and process a file using set of Activities that run on the same host. Activities are executed to download a
  file from the web, store it locally on the host, and then "process it". This samples showcases how to handle a
  scenario where all subsequent Activities need to execute on the same host as the first Activity in the sequence. In
  Go, this is achieved by using the Session APIs.

- [**Particle Swarm Optimization**](https://github.com/temporalio/samples-go/tree/main/pso): Demonstrates how to
  perform a long iterative math optimization process using particle swarm optimization (PSO). This sample showcases the
  use of parallel executions, `ContinueAsNew` for long histories, a Query API, and the use of a custom `DataConverter`
  for serialization.

- [**Prometheus Metrics**](https://github.com/temporalio/samples-go/tree/main/metrics): Demonstrates how to instrument
  Temporal with Prometheus and Uber's Tally library.

- [**Request/Response with Response Activities**](https://github.com/temporalio/samples-go/tree/main/reqrespactivity):
  Demonstrates how to accept requests via signals and use callback activities to push responses.

- [**Request/Response with Response Queries**](https://github.com/temporalio/samples-go/tree/main/reqrespquery):
  Demonstrates how to accept requests via signals and use queries to poll for responses.

### Pending examples

Mostly examples we haven't yet ported from https://github.com/temporalio/samples-java/

- Async activity calling: *Example to be completed*
- Async lambda:  *Example to be completed*
- Periodic Workflow: Workflow that executes some logic periodically. *Example to be completed*
- Exception propagation and wrapping: *Example to be completed*
- Polymorphic activity: *Example to be completed*
- Side Effect:  *Example to be completed* - [Docs](https://docs.temporal.io/go/how-to-execute-a-side-effect-in-go)

### Fixtures

These are edge case examples useful for Temporal internal development and bug
reporting. [See their readme for more details](https://github.com/temporalio/samples-go/tree/main/temporal-fixtures).
