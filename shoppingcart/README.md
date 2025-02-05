# Shopping Cart

This sample workflow shows how a shopping cart application can be implemented.
This sample utilizes Update-with-Start and the `WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING`
option to start and continually update the workflow with the same Update-with-Start
call. This is also known as lazy-init. You will see in the Temporal UI, when you checkout
your cart, the workflow will complete and `ui/main.go` will throw an error 
`workflow execution already completed`. This example can be extended to handle concurrent
shoppers (would need some sort of SessionID) or support starting a new session/workflow
after checkout.

Another interesting Update-with-Start use case is 
[early return](https://github.com/temporalio/samples-go/tree/main/early-return), 
which supplements this sample and can be used to handle the transaction and payment
portion of this shopping cart scenario.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).

    NOTE: frontend.enableExecuteMultiOperation=true must be configured for the server
in order to use Update-with-Start. For example:
```
temporal server start-dev --dynamic-config-value frontend.enableExecuteMultiOperation=true
```

2) Run the following command to start the worker
```
go run shoppingcart/worker/main.go
```
3) Run the following command to start the web app
```
go run shoppingcart/ui/main.go
```
