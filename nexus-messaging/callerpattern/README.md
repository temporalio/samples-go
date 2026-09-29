## Caller pattern

The handler worker starts a `GreetingWorkflow` for a User ID.
The Nexus handler holds that ID and routes every Nexus operation to it.
The caller’s input doesn’t include the Workflow ID because it isn’t known. Instead, the caller provides the User ID, and the handler derives the Workflow ID from it (see `GetWorkflowID`).

The handler worker uses the same `GetWorkflowID` call to generate a Workflow ID from a User ID when it launches the Workflow.

The caller Workflow:
1. Queries for supported languages (`getLanguages` -- backed by a query handler)
2. Changes the language to Arabic (`setLanguage` -- backed by an update handler that calls an activity)
3. Confirms the change via a second query (`getLanguage`)
4. Approves the Workflow (`approve` -- backed by a signal handler)

### Running

Start a Temporal server:

```bash
temporal server start-dev
```

Create the namespaces and Nexus endpoint:

```bash
temporal operator namespace create --namespace my-target-namespace
temporal operator namespace create --namespace my-caller-namespace

temporal operator nexus endpoint create \
  --name my-nexus-endpoint-name \
  --target-namespace my-target-namespace \
  --target-task-queue my-handler-task-queue
```

In one terminal, start the handler worker:

```bash
go run ./nexus-messaging/callerpattern/handler/worker/main.go
```

In a second terminal, start the caller worker:

```bash
go run ./nexus-messaging/callerpattern/caller/worker/main.go
```

In a third terminal, run the following command to start the example

```bash
go run ./nexus-messaging/callerpattern/caller/starter/main.go
```

Expected output:

```
[1] getLanguages returned 2 languages
[2] setLanguage(Arabic) returned previous language: English
[3] getLanguage returned: Arabic (confirmed Arabic)
[4] approve sent successfully
```
