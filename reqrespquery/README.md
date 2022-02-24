# Request/Response Sample with Query-Based Responses

This sample demonstrates how to send a request and get a response from a Temporal workflow via a query.

The workflow in this specific example accepts requests to uppercase a string via signal and then provides the response
via a query. This means the requester must poll for response via queries.

### Running

Follow the below steps to run this sample:

1) You need a Temporal service running. See details in README.md.

2) Run the following command in the background or in another terminal to run the worker:

    go run ./reqrespquery/worker

3) Run the following command to start the workflow:

    go run ./reqrespquery/starter

4) Run the following command to uppercase a string every second:

    go run ./reqrespquery/request

Multiple of those can be run on different terminals to confirm that the processes are independent.

### Comparison with activity-based responses

In the [reqrespactivity](../reqrespactivity) sample, we use activities to send back responses. Here are the pros/cons of
this approach compared to activity-based responses:

Pros:

* Query-based responses don't require a worker on the caller side
* Query-based responses do not have to record the response/query in history
* Query-based responses can occur even after the workflow has been terminated

Cons:

* Query-based responses are often slower due to polling instead of pushing
* Query-based responses require the workflow to store the response state explicitly
* Query-based responses cannot know on the workflow side whether a response was received

### Explanation of continue-as-new

Workflows cannot have infinitely-sized history and when the event count grows too large, `ContinueAsNew` can be returned
to start a new one atomically. However, in order not to lose any data, signals must be drained and any other futures
that need to be reacted to must be completed first. This means there must be a period where there are no signals to
drain and no futures to wait on. If signals come in faster than processed or futures wait so long there is no idle
period, `ContinueAsNew` will not happen in a timely manner and history will grow.

Since this sample is a long-running workflow, once the request count reaches a certain size, we perform a
`ContinueAsNew`. To not lose any data, we only send this if there are no in-flight signal requests or executing
activities. An executing activity can mean it is busy retrying. Care must be taken designing these systems where they do
not receive requests so frequent that they can never have a idle period to return a `ContinueAsNew`. Signals are usually
fast to receive, so they are less of a problem. Waiting on activities (as a response to the request or as a response
callback activity) can be a tougher problem. Since we rely on the response of the activity, activity must complete
including all retries. Retry policies of these activities should be set balancing the resiliency needs with the need to
have a period of idleness at some point.
