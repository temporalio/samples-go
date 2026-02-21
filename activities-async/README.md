# Async Activities Invocation in GoLang

This sample shows how to submit two activities in non-blocking fashion and wait until both are completed to return,
or fail if one of the two fails. The example uses two activity invocation, registers that on a selector, and invoke
`selector.Select(ctx)` twice, since each selector call will proceed as soon as one result is available.



### Running this sample

```bash
go run activities-async/worker/main.go
```

Start the Workflow Execution:

```bash
go run activities-async/starter/main.go
```
