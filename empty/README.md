<!-- @@@SNIPSTART samples-go-empty-readme -->
Make sure the [Temporal Server is running locally](https://docs.temporal.io/docs/server/quick-install).

From the root of the project, start a Worker:

```bash
go run empty/worker/main.go
```

Start the Workflow Execution:

```bash
go run empty/starter/main.go
```
<!-- @@@SNIPEND -->
