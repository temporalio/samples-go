<!-- @@@SNIPSTART samples-go-cancellation-readme -->
Make sure the [Temporal Server is running locally](https://docs.temporal.io/docs/server/quick-install).

From the root of the project, start a Worker:

```bash
go run cancellation/worker/main.go
```

Start the Workflow Execution:

```bash
go run cancellation/starter/main.go
```

Cancel the Workflow Execution:

```bash
go run cancellation/cancel/main.go
```
<!-- @@@SNIPEND -->
