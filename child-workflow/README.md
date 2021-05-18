<!--@@@SNIPSTART samples-go-child-workflow-example-readme-->
### Run sample

1. 1. Make sure the [Temporal Server is running locally](https://docs.temporal.io/docs/server/quick-install).

2. From the root of the project, start the Worker:

```
go run child-workflow/worker/main.go
```

3. Start the Workflow Execution

```
go run child-workflow/starter/main.go
```
<!--@@@SNIPEND-->
