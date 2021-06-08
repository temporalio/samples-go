# Temporal Go SDK Samples

[![FOSSA Status](https://app.fossa.com/api/projects/custom%2B18405%2Fgithub.com%2Ftemporalio%2Fsamples-go.svg?type=shield)](https://app.fossa.com/projects/custom%2B18405%2Fgithub.com%2Ftemporalio%2Fsamples-go?ref=badge_shield)

This repository contains several sample Workflow applications that demonstrate the various capabilities of the Temporal Server via the Temporal Go SDK.

- Temporal Server repo: [https://github.com/temporalio/temporal](https://github.com/temporalio/temporal)
- Temporal Go SDK repo: [https://github.com/temporalio/sdk-go](https://github.com/temporalio/sdk-go)
- Go SDK docs: https://docs.temporal.io/docs/go/introduction

## How to use

Make sure the Temporal Server is running locally.
Follow the [Quick install guide](https://docs.temporal.io/docs/server/quick-install) to do that.

Then, clone this repository and follow the instructions in the README that is included with each sample.
The [helloworld](helloworld/README.md) sample is a good place to start.

You can learn more about running the Server locally in the [temporalio/docker-compose README](https://github.com/temporalio/docker-compose/blob/main/README.md).
And you can learn more about the Temporal Server technology in general via our [documentation](https://docs.temporal.io/).

## Samples Directory

Each sample demonstrates one feature of the SDK, together with tests.

- [Hello World](https://github.com/temporalio/samples-go/tree/master/helloworld): Simple example of a Workflow and an Activity.

### API demonstrations

  - Async activity completion: Can be observed in the [Expense Report](https://github.com/temporalio/samples-go/tree/master/expense) example. [Docs](https://docs.temporal.io/docs/go/activities#asynchronous-activity-completion)
  - [Retry Activity](https://github.com/temporalio/samples-go/tree/master/retryactivity): Executes an unreliable activity **with retry policy**. If activity execution failed, server will schedule retry based on retry policy configuration. The activity also **heartbeats progress** so it could resume from reported progress in retry attempt.
  - [Child Workflow](https://github.com/temporalio/samples-go/tree/master/child-workflow): Demonstrates how to use invoke child workflow from parent workflow execution.  Each child workflow execution is starting a new run and parent execution is notified only after the completion of last run.
    - [Child Workflow with ContinueAsNew](https://github.com/temporalio/samples-go/tree/master/child-workflow-continue-as-new): Demonstrates that a child workflow calling continue as new is *not visible by a parent*. Parent receives notification about a child completion only when a child completes, fails or times out. This is a useful feature when there is a need to **process a large set of data**. The child can iterate over the data set calling continue as new periodically without polluting the parents' history.
  - [Cancellation](https://github.com/temporalio/samples-go/tree/master/cancelactivity): How to cancel a running workflow with `CancelWorkflow` and defer a cleanup activity for execution after cancel
  - Coroutines: You should not use native `go` routines - use Temporal coroutines `workflow.Go()` instead ([for determinism](https://docs.temporal.io/docs/go/workflows/#how-to-write-workflow-code)). Seen in: Split/Merge, DSL, Recovery, PSO, Parallel Workflow examples.
  - [Cron Workflow](https://github.com/temporalio/samples-go/tree/master/cron): Recurring workflow that is executed according to a cron schedule. Uses `HasLastCompletionResult` and `GetLastCompletionResult` to pass information between runs. [Docs](https://docs.temporal.io/docs/go/distributed-cron/)
  - [Encrypted Payloads](https://github.com/temporalio/samples-go/tree/master/encrypted-payloads): How to customize encryption/decryption of Workflow data with the DataConverter API. [Docs](https://docs.temporal.io/docs/go/workflows/#custom-serialization-and-workflow-security).
    - [Crypt Converter](https://github.com/temporalio/samples-go/tree/master/cryptconverter): Advanced, newer example.
  - [Query Example](https://github.com/temporalio/samples-go/tree/master/query): Demonstrates how to query a state of a single workflow using `QueryWorkflow` and `SetQueryHandler`. [Docs](https://docs.temporal.io/docs/go/queries)
  - Selectors: You should not use native Go `select` - use [Go SDK Selectors](https://docs.temporal.io/docs/go/selectors) `selector.Select(ctx)` instead ([for determinism](https://docs.temporal.io/docs/go/workflows/#how-to-write-workflow-code)). Seen in: Pick First, Mutex, DSL, and Timer examples.
  - Sessions: used in [File Processing example](https://github.com/temporalio/samples-go/tree/master/fileprocessing). [Docs](https://docs.temporal.io/docs/go/sessions).
  - Signals: used in [Recovery](https://github.com/temporalio/samples-go/tree/master/recovery) and [Mutex](https://github.com/temporalio/samples-go/tree/master/mutex) examples. See [ecommerce app blogpost](https://docs.temporal.io/blog/build-an-ecommerce-app-with-temporal-part-1) and [Docs](https://docs.temporal.io/docs/go/signals).
  - [Search Attributes](https://github.com/temporalio/samples-go/tree/master/searchattributes): Custom search attributes that can be used to find workflows using predicates (must use with Elasticsearch)
  - [Timer Futures](https://github.com/temporalio/samples-go/tree/master/timer): Starts a long running order processing operation and in the case that the processing takes too long, we want to send out a notification email to user about the delay, but we won't cancel the operation. If the operation finishes before the timer fires, then we want to cancel the timer. Using `workflow.NewTimer`
  - [Tracing and Context Propagation](https://github.com/temporalio/samples-go/tree/master/ctxpropagation): Initializes the client with a context propagator which propagates specific information in the `context.Context` object across the Workflow. The `context.Context` object is populated with the information prior to calling `StartWorkflow`. The Workflow demonstrates that the information is available in the Workflow and any activities executed. [Docs](https://docs.temporal.io/docs/go/tracing/).

### Dynamic Workflow Logic Examples

Demonstrating common control flow patterns used together with Temporal's API

  - [Dynamic Invocation](https://github.com/temporalio/samples-go/tree/master/dynamic): Demonstrates invocation of workflows and activities using name rather than strongly typed function.
  - [Branching Workflow](https://github.com/temporalio/samples-go/blob/master/branch): Executes multiple activities in parallel. The number of branches is controlled by a passed in parameter.
  - [Exclusive Choice](https://github.com/temporalio/samples-go/tree/master/choice-exclusive): How to execute activities based on dynamic input
  - [Multi-Choice](https://github.com/temporalio/samples-go/tree/master/choice-multi): Demonstrates how to run multiple activities in parallel based on a dynamic input.
  - [Mutex workflow](https://github.com/temporalio/samples-go/tree/master/mutex): Demos an ability to lock/unlock a particular resource within a particular Temporal namespace so that other workflows within the same namespace would wait until a resource lock is released. This is useful when we want to avoid race conditions or parallel mutually exclusive operations on the same resource.
  - [Parallel workflow](https://github.com/temporalio/samples-go/tree/master/parallel): Executes multiple branches in parallel using `workflow.Go()` method.
  - [Pick First](https://github.com/temporalio/samples-go/tree/master/pickfirst): Execute activities in parallel branches, pick the result of the branch that completes first, and then cancels other activities that are not finished yet.
  - [Split/Merge](https://github.com/temporalio/samples-go/tree/master/splitmerge): Demonstrates how to use multiple Temporal coroutines (instead of native goroutine) to process a chunk of a large work item in parallel, and then merge the intermediate result to generate the final result. In Temporal workflow, you should not use go routine. Instead, you use corotinue via workflow.Go method.
  - [Synchronous Proxy workflow](https://github.com/temporalio/samples-go/tree/master/synchronous-proxy) pattern: Achieve synchronous interaction with a main workflow from a "proxy workflow". The proxy workflow sends a signal to the main workflow, then blocks waiting for a signal in response.

### Concrete Examples

  - [DSL Workflow](https://github.com/temporalio/samples-go/tree/master/dsl): Demonstrates how to implement a DSL workflow. In this sample, we provide 2 sample yaml files each defines a custom workflow that can be processed by this DSL workflow sample code. Useful if you want to build your own low code
  - [Expense Request](https://github.com/temporalio/samples-go/tree/master/expense): Process an expense request. The key part of this sample is to show how to complete an activity asynchronously.
  - [File Processing](https://github.com/temporalio/samples-go/tree/master/fileprocessing): Demos a file processing process. The workflow first starts an activity to download a requested resource file from web and store it locally on the host where it runs the download activity. Then, the workflow will start more activities to process the downloaded resource file. The key part is the following activities have to be run on the same host as the initial downloading activity. This is achieved by using the session API.
  - [Particle Swarm Optimization](https://github.com/temporalio/samples-go/tree/master/pso): Demos a long iterative math optimization process using particle swarm optimization (PSO). It demonstrates usage of parallel execution, `ContinueAsNew` for long histories, a query API, and custom `DataConverter` serialization.
  - [Prometheus Metrics](https://github.com/temporalio/samples-go/tree/master/metrics): Demonstrates how to instrument Temporal with Prometheus and Uber's Tally library.

### Misc/Pending Examples

Mostly examples we haven't yet ported from https://github.com/temporalio/samples-java/

  - Async activity calling: *Example to be completed*
  - Async lambda:  *Example to be completed*
  - Periodic Workflow: Workflow that executes some logic periodically. *Example to be completed*
  - Exception propagation and wrapping: *Example to be completed*
  - Polymorphic activity: *Example to be completed*
  - SAGA pattern:  *Example to be completed*
  - Side Effect:  *Example to be completed* - [Docs](https://docs.temporal.io/docs/go/side-effect)

### Fixtures

These are edge case examples useful for Temporal internal development and bug reporting. [See their readme for more details](https://github.com/temporalio/samples-go/tree/master/temporal-fixtures).
