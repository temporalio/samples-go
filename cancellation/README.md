<!--@@@SNIPSTART samples-go-cancellation-readme-->
### Run sample

1. Make sure the [Temporal Server is running locally](https://docs.temporal.io/docs/server/quick-install).

2. From the root of the project, start the Worker:

```bash
go run cancellation/worker/main.go
```

3. Start the Workflow Execution:

```bash
go run cancellation/starter/main.go
```

4. Cancel the Workflow Execution:

```bash
go run cancellation/cancel/main.go
```
<!--@@@SNIPEND-->
