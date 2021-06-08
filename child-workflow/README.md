## Child Workflow
<!-- @@@SNIPSTART samples-go-child-workflow-example-readme -->
Make sure the [Temporal Server is running locally](https://docs.temporal.io/docs/server/quick-install).

From the root of the project, start a Worker:

```bash
go run child-workflow/worker/main.go
```

Start the Workflow Execution:

```bash
go run child-workflow/starter/main.go
```
<!-- @@@SNIPEND -->
