# Request/Response Sample with Update-Based Responses

This sample demonstrates how to send a request and get a response from a Temporal workflow via an update.

[Update](https://docs.temporal.io/workflows#update) is a new feature available for preview on [Temporal Server v1.21](https://github.com/temporalio/temporal/releases/tag/v1.21.0).

### Running

Follow the below steps to run this sample:

1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use). Note, update is only supported on [Temporal Server v1.21](https://github.com/temporalio/temporal/releases/tag/v1.21.0)

2) Update functionality is disabled by default but can be enabled per Namespace by setting the `frontend.enableUpdateWorkflowExecution` flag to true for that Namespace in dynamic config.

3) Run the following command in the background or in another terminal to run the worker:

    go run ./reqrespupdate/worker

4) Run the following command to start the workflow:

    go run ./reqrespupdate/starter

5) Run the following command to uppercase a string every second:

    go run ./reqrespupdate/request

Multiple of those can be run on different terminals to confirm that the processes are independent.

### Comparison with activity-based responses and query-based response

There are two other samples showing how to do a request response pattern under [reqrespactivity](../reqrespactivity) sample and the [reqrespquery](../reqrespquery). The update based approach is the superior option, and once released, will be the recommend approach.


### Explanation of continue-as-new

Workflows cannot have infinitely-sized history and when the event count grows too large, `ContinueAsNew` can be returned
to start a new one atomically. However, in order not to lose any data, update requests must be handled and any other futures
that need to be reacted to must be completed first. This means there must be a period where there are no updates to
process and no futures to wait on. If update requests come in faster than processed or futures wait so long there is no idle
period, `ContinueAsNew` will not happen in a timely manner and history will grow.

Since this sample is a long-running workflow, once the request count reaches a certain size, we perform a
`ContinueAsNew`. To not lose any data, we only send this if there are no in-flight update requests or executing
activities. An executing activity can mean it is busy retrying. Care must be taken designing these systems where they do
not receive requests so frequent that they can never have a idle period to return a `ContinueAsNew`. Signals are usually
fast to receive, so they are less of a problem. Waiting on activities (as a response to the request or as a response
callback activity) can be a tougher problem. Since we rely on the response of the activity, activity must complete
including all retries. Retry policies of these activities should be set balancing the resiliency needs with the need to
have a period of idleness at some point.
