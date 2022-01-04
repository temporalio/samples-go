# Request/Response Sample

This sample demonstrates how to send a request and get a response from a Temporal workflow. Two approaches to providing
the caller a response are in the example:

* Query poll - Caller can run queries polling for response (the less efficient way)
* Response activity - Caller can provide activity information on where to send the response to (the more efficient way)

The workflow in this specific example accepts requests to uppercase a string via signal and then provides the response
via either approach above.

### Running

Follow the below steps to run this sample:

1) You need a Temporal service running. See details in README.md.

2) Run the following command in the background or in another terminal to run the worker:

    go run ./reqresp/worker

3) Run the following command to start the workflow:

    go run ./reqresp/starter

4) Run the following command to uppercase some strings via the slower query-response approach:

    go run ./reqresp/request foo bar

Among other log lines, this will output the following after a few seconds:

    Uppercased - from foo to FOO
    Uppercased - from bar to BAR

5) Run the following commands to uppercase some strings via the faster activity-response approach:

    go run ./reqresp/request -activity-based foo bar

This will output the same thing, but much faster.

### Explanation of response handling

When a query is executed in Temporal, it must go to the worker, invoke the query handler (potentially replaying if the
workflow is not already present/cached), and return the query back. Therefore, in order to do a query-based
request/response, one must signal the workflow and continuously query the workflow for the response. The query can't be
executed too frequently as this affects server and worker resources. This is a method that, while simple, is much
slower.

Alternatively, a Temporal workflow can execute activities on different task queues. Therefore, a "response" task queue
can be setup to handle activities executed from the other workflow. This means the response can be pushed. By providing
which activity and task queue the response should be sent to inside the request, the requester can just be notified when
the activity is executed.

In this particular example, we have abstracted both concepts out into a `Requester`. This requester can be reused
although the sample just shows it used ephemerally as part of the `request` CLI.

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