# Shopping Cart

This sample workflow shows how a shopping cart application can be implemented.
Note that this program uses websockets to communicate between the webapp and 
the Temporal service.

The shopping cart is represented as a workflow, maintaining the state of the
cart, and the web socket server updates the carts with signals, and retrieves
the cart state with a query. See [workflow message passing](https://docs.temporal.io/encyclopedia/workflow-message-passing)
on the difference between queries and signals.

### Steps to run this sample:
1) Run a [Temporal service](https://github.com/temporalio/samples-go/tree/main/#how-to-use).
2) Run the following command to start the worker
```
go run shoppingcart/worker/main.go
```
3) Run the following command to start the web socket server
```
go run shoppingcart/websocket/main.go
```
4) Run the following command to start the web app
```
go run shoppingcart/webapp/main.go
```
5) Run the following command to start the workflow execution
```
go run shoppingcart/starter/main.go
```