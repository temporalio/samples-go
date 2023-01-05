## Parallel Activities
<!-- @@@SNIPSTART samples-go-branch-readme -->
Make sure the [Temporal Server is running locally](https://docs.temporal.io/clusters/quick-install).

From the root of the project, start a Worker:

```bash
go run branch/worker/main.go
```

Start the Workflow Execution:

```bash
go run branch/starter/main.go
```
<!-- @@@SNIPEND -->
