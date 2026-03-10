## Cancellation
<!-- @@@SNIPSTART samples-go-cancellation-readme -->
Make sure the [Temporal Server is running locally](https://learn.temporal.io/getting_started/go/dev_environment/#set-up-a-local-temporal-service-for-development-with-temporal-cli).

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
