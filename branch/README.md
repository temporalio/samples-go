<!--@@@SNIPSTART samples-go-branch-readme-->
### Run sample

1. Make sure the [Temporal Server is running locally](https://docs.temporal.io/docs/server/quick-install).

2. From the root of the project, start a Worker:

```bash
go run branch/worker/main.go
```

3. Start the Workflow Execution

```bash
go run branch/starter/main.go
```
<!--@@@SNIPEND-->
