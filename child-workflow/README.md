## Child Workflow
<!-- @@@SNIPSTART samples-go-child-workflow-example-readme -->
Make sure the [Temporal Server is running locally](https://learn.temporal.io/getting_started/go/dev_environment/#set-up-a-local-temporal-service-for-development-with-temporal-cli).

From the root of the project, start a Worker:

```bash
go run child-workflow/worker/main.go
```

Start the Workflow Execution:

```bash
go run child-workflow/starter/main.go
```
<!-- @@@SNIPEND -->
