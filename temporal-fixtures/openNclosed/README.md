This fixuture creates 20 workflows - all either open for 10 minutes, or all either closed right away, depending on the `keep` flag:

1. `keep` argument makes the workflow remain Open for 10 minutes if set to `true`, or otherwise the workflows will be quickly Close.
2. Change the `Namespace: "default"` to another namespace for testing with a new set of data/runs 

Used in:

- https://github.com/temporalio/web/pull/315
