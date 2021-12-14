# Request/Response Sample

This sample demonstrates how to send a request and get a response from a Temporal workflow. Two approaches to providing
the caller a response are in the example:

* Query poll - Caller can run queries polling for response (the slow/inefficient way)
* Response activity - Caller can provide activity information on where to send the response to (the more robust way)

The workflow in this specific example accepts requests to uppercase a string via signal and then provides the response
via either approach above. Care is taken to continue-as-new the workflow when the request count gets too large, but like
all uses of continue-as-new in Go, the signal must be drained first. Therefore, if the requests come in faster than they
are processed, the workflow history can grow until it reaches too large of a size and is then terminated.

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
which activity and task queue the response should be sent to inside the request, the local requester can just be
notified when the local activity is executed.

In this particular example, we have abstracted both concepts out into a `Requester`. This requester can be reused
although the sample just shows it used ephemerally as part of the `request` CLI.